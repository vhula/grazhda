package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/charmbracelet/log"
	"github.com/vhula/grazhda/internal/config"
	pb "github.com/vhula/grazhda/internal/proto"

	"google.golang.org/grpc"
)

type workspaceServer struct {
	pb.UnimplementedWorkspaceServiceServer
	pb.UnimplementedDukhServiceServer
	workspaces map[string]*pb.Workspace
	config     *config.Config
	grpcServer *grpc.Server
	pidFile    string
	mu         sync.RWMutex
}

func (s *workspaceServer) InitWorkspaces(ctx context.Context, req *pb.InitWorkspacesRequest) (*pb.InitWorkspacesResponse, error) {
	_ = ctx
	_ = req
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Info("Init workspaces start")
	var statuses []string
	for _, wsConfig := range s.config.Workspaces {
		// create workspace dir
		err := os.MkdirAll(wsConfig.Path, 0755)
		if err != nil {
			msg := fmt.Sprintf("failed to create dir for %s: %v", wsConfig.Name, err)
			statuses = append(statuses, msg)
			log.Error(msg)
			continue
		}
		// create log file for workspace
		logPath := filepath.Join(wsConfig.Path, "dukh.log")
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			msg := fmt.Sprintf("failed to create log for %s: %v", wsConfig.Name, err)
			statuses = append(statuses, msg)
			log.Error(msg)
			continue
		}
		defer file.Close()
		createdAt := time.Now().Format(time.RFC3339)
		_, err = fmt.Fprintf(file, "Id: %s, Name: %s, CreatedAt: %s\n", wsConfig.Name, wsConfig.Name, createdAt)
		if err != nil {
			msg := fmt.Sprintf("failed to write log for %s: %v", wsConfig.Name, err)
			statuses = append(statuses, msg)
			log.Error(msg)
			continue
		}
		// add workspace to map
		s.workspaces[wsConfig.Name] = &pb.Workspace{
			Id:           wsConfig.Name,
			Name:         wsConfig.Name,
			CreatedAt:    createdAt,
			AbsolutePath: wsConfig.Path,
		}
		msg := fmt.Sprintf("initialized workspace %s", wsConfig.Name)
		statuses = append(statuses, msg)
		log.Info(msg)

		// create project directories and clone repositories
		for _, project := range wsConfig.Projects {
			projectDir := filepath.Join(wsConfig.Path, project.Name)
			err := os.MkdirAll(projectDir, 0755)
			if err != nil {
				msg := fmt.Sprintf("failed to create project dir for %s/%s: %v", wsConfig.Name, project.Name, err)
				statuses = append(statuses, msg)
				log.Error(msg)
				continue
			}
			msg := fmt.Sprintf("created project directory %s", projectDir)
			log.Info(msg)

			// handle subprojects or direct repositories
			if len(project.Subprojects) > 0 {
				// iterate over subprojects
				for _, subproject := range project.Subprojects {
					for _, repo := range subproject.Repositories {
						gitCmd := constructGitCommand(wsConfig.CloneCommandTemplate, repo.Name, projectDir, subproject.Branch, repo.LocalDirName)
						log.Info(fmt.Sprintf("executing: %s", gitCmd))
						err := executeCommand(gitCmd, projectDir)
						if err != nil {
							msg := fmt.Sprintf("failed to execute git clone for %s: %v", repo.Name, err)
							statuses = append(statuses, msg)
							log.Error(msg)
							continue
						}
						msg := fmt.Sprintf("cloned %s successfully", repo.Name)
						statuses = append(statuses, msg)
						log.Info(msg)
					}
				}
			} else if len(project.Repositories) > 0 {
				// iterate over direct repositories
				for _, repo := range project.Repositories {
					gitCmd := constructGitCommand(wsConfig.CloneCommandTemplate, repo.Name, projectDir, project.Branch, repo.LocalDirName)
					log.Info(fmt.Sprintf("executing: %s", gitCmd))
					err := executeCommand(gitCmd, projectDir)
					if err != nil {
						msg := fmt.Sprintf("failed to execute git clone for %s: %v", repo.Name, err)
						statuses = append(statuses, msg)
						log.Error(msg)
						continue
					}
					msg := fmt.Sprintf("cloned %s successfully", repo.Name)
					statuses = append(statuses, msg)
					log.Info(msg)
				}
			}
		}
	}
	log.Info("Init workspaces end")
	return &pb.InitWorkspacesResponse{Statuses: statuses}, nil
}

func (s *workspaceServer) PurgeWorkspaces(ctx context.Context, req *pb.PurgeWorkspacesRequest) (*pb.PurgeWorkspacesResponse, error) {
	_ = ctx
	_ = req
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Info("Purge workspaces start")
	var statuses []string
	for _, wsConfig := range s.config.Workspaces {
		if _, exists := s.workspaces[wsConfig.Name]; exists {
			delete(s.workspaces, wsConfig.Name)
		}
		// delete directory
		err := os.RemoveAll(wsConfig.Path)
		if err != nil {
			msg := fmt.Sprintf("failed to delete directory for %s: %v", wsConfig.Name, err)
			statuses = append(statuses, msg)
			log.Error(msg)
			continue
		}
		msg := fmt.Sprintf("purged %s at %s", wsConfig.Name, wsConfig.Path)
		statuses = append(statuses, msg)
		log.Info(msg)
	}
	log.Info("Purge workspaces end")
	return &pb.PurgeWorkspacesResponse{Statuses: statuses}, nil
}

func (s *workspaceServer) GetWorkspaces(ctx context.Context, req *pb.GetWorkspacesRequest) (*pb.GetWorkspacesResponse, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	log.Info("Get workspaces start")
	var workspaces []*pb.Workspace
	for _, id := range req.Ids {
		if ws, exists := s.workspaces[id]; exists {
			workspaces = append(workspaces, ws)
		}
	}
	log.Info("Get workspaces end")

	return &pb.GetWorkspacesResponse{Workspaces: workspaces}, nil
}

func (s *workspaceServer) StopDukh(ctx context.Context, req *pb.StopDukhRequest) (*pb.StopDukhResponse, error) {
	_ = ctx
	_ = req

	log.Info("Received gRPC stop request for Dukh server")

	go func() {
		time.Sleep(100 * time.Millisecond)
		if s.pidFile != "" {
			_ = os.Remove(s.pidFile)
		}
		s.grpcServer.GracefulStop()
	}()

	return &pb.StopDukhResponse{Status: "dukh server stopping"}, nil
}

func (s *workspaceServer) StatusDukh(ctx context.Context, req *pb.StatusDukhRequest) (*pb.StatusDukhResponse, error) {
	_ = ctx
	_ = req

	return &pb.StatusDukhResponse{
		Running: true,
		Pid:     int32(os.Getpid()),
		Status:  "dukh server is running",
	}, nil
}

func startServer(cfg *config.Config) {
	pidFilePath, err := getDukhPIDFilePath()
	if err != nil {
		log.Fatalf("Failed to resolve pid file path: %v", err)
	}

	err = os.WriteFile(pidFilePath, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0644)
	if err != nil {
		log.Fatalf("Failed to write pid file: %v", err)
	}
	defer os.Remove(pidFilePath)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Dukh.Host, cfg.Dukh.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := &workspaceServer{
		workspaces: make(map[string]*pb.Workspace),
		config:     cfg,
		pidFile:    pidFilePath,
	}

	s := grpc.NewServer()
	server.grpcServer = s
	pb.RegisterWorkspaceServiceServer(s, server)
	pb.RegisterDukhServiceServer(s, server)

	log.Printf("Dukh gRPC server starting on %s:%d", cfg.Dukh.Host, cfg.Dukh.Port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func main() {
	err, _ := run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(args []string) (error, *config.Config) {
	if len(args) < 2 {
		fmt.Println("Dukh - The Worker CLI")
		fmt.Println("Usage: dukh <command>")
		fmt.Println("Commands:")
		fmt.Println("  start    - Start the Dukh gRPC server")
		fmt.Println("  stop     - Stop the Dukh server")
		fmt.Println("  status   - Check server status")
		return nil, nil
	}

	command := args[1]
	switch command {
	case "start":
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err), nil
		}
		fmt.Println("Starting Dukh gRPC server...")
		startServer(cfg)
		return nil, nil
	case "stop":
		fmt.Println("Stopping Dukh server...")
		if err := stopDukhServer(); err != nil {
			return fmt.Errorf("failed to stop dukh server: %w", err), nil
		}
		return nil, nil
	case "status":
		fmt.Println("Dukh server status: Not implemented yet")
		return nil, nil
	default:
		return fmt.Errorf("unknown command: %s", command), nil
	}
}

func getDukhPIDFilePath() (string, error) {
	grazhdaDir := os.Getenv("GRAZHDA_DIR")
	if grazhdaDir == "" {
		return "", fmt.Errorf("GRAZHDA_DIR is not set")
	}
	return filepath.Join(grazhdaDir, "dukh.pid"), nil
}

func stopDukhServer() error {
	pidFilePath, err := getDukhPIDFilePath()
	if err != nil {
		return err
	}

	pidBytes, err := os.ReadFile(pidFilePath)
	if err != nil {
		return fmt.Errorf("unable to read pid file %s: %w", pidFilePath, err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return fmt.Errorf("invalid pid in %s: %w", pidFilePath, err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("unable to find process %d: %w", pid, err)
	}

	if err := process.Kill(); err != nil {
		return fmt.Errorf("unable to stop process %d: %w", pid, err)
	}

	if err := os.Remove(pidFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to remove pid file %s: %w", pidFilePath, err)
	}

	log.Info("Dukh server stopped", "pid", pid)
	return nil
}

func constructGitCommand(tmplStr, repoName, projectDir, branch, localDirName string) string {
	// Calculate DestDir: use localDirName if provided, otherwise use repoName
	destDir := localDirName
	if destDir == "" {
		destDir = repoName
	}

	// Create template data structure
	data := struct {
		Repository  string
		RepoName    string
		DestDir     string
		Destination string
		Branch      string
	}{
		Repository:  repoName,
		RepoName:    repoName,
		DestDir:     destDir,
		Destination: filepath.Join(projectDir, destDir),
		Branch:      branch,
	}

	// Parse and execute template
	tmpl, err := template.New("git").Parse(tmplStr)
	if err != nil {
		log.Error(fmt.Sprintf("failed to parse template: %v", err))
		return ""
	}

	var result string
	buf := &strings.Builder{}
	err = tmpl.Execute(buf, data)
	if err != nil {
		log.Error(fmt.Sprintf("failed to execute template: %v", err))
		return ""
	}

	result = buf.String()
	return result
}

func executeCommand(command string, dir string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
