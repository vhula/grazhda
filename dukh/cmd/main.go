package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/vhula/grazhda/internal/config"
	pb "github.com/vhula/grazhda/internal/proto"

	"google.golang.org/grpc"
)

type dukhServer struct {
	pb.UnimplementedDukhServiceServer
	grpcServer *grpc.Server
	pidFile    string
}

func (s *dukhServer) StopDukh(ctx context.Context, req *pb.StopDukhRequest) (*pb.StopDukhResponse, error) {
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

func (s *dukhServer) StatusDukh(ctx context.Context, req *pb.StatusDukhRequest) (*pb.StatusDukhResponse, error) {
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

	server := &dukhServer{pidFile: pidFilePath}

	s := grpc.NewServer()
	server.grpcServer = s
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
