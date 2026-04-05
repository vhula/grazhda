package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	dukhpb "github.com/vhula/grazhda/dukh/proto"
	"google.golang.org/grpc"
)

const version = "0.1.0"

// Server is the dukh gRPC server.
type Server struct {
	dukhpb.UnimplementedDukhServiceServer

	grpcServer *grpc.Server
	monitor    *Monitor
	logger     *log.Logger
	startedAt  time.Time
	stopped    atomic.Bool
}

// New creates a Server. The monitor must already be running (call Start separately).
func New(monitor *Monitor, logger *log.Logger) *Server {
	return &Server{
		monitor:   monitor,
		logger:    logger,
		startedAt: time.Now(),
	}
}

// ListenAndServe starts the gRPC server on addr and blocks until stopped.
func (s *Server) ListenAndServe(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	s.grpcServer = grpc.NewServer()
	dukhpb.RegisterDukhServiceServer(s.grpcServer, s)

	s.logger.Info("dukh: gRPC server listening", "addr", addr)
	return s.grpcServer.Serve(lis)
}

// GracefulStop signals the server to stop accepting new requests and finish in-flight ones.
func (s *Server) GracefulStop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

// Stop implements DukhService.Stop — triggers a graceful shutdown.
func (s *Server) Stop(_ context.Context, _ *dukhpb.StopRequest) (*dukhpb.StopResponse, error) {
	s.logger.Info("dukh: Stop RPC received — initiating graceful shutdown")
	// Trigger shutdown asynchronously so the RPC response can be sent first.
	go func() {
		time.Sleep(100 * time.Millisecond)
		s.GracefulStop()
	}()
	return &dukhpb.StopResponse{Message: "dukh shutting down"}, nil
}

// Status implements DukhService.Status — returns current workspace health.
func (s *Server) Status(_ context.Context, req *dukhpb.StatusRequest) (*dukhpb.StatusResponse, error) {
	snapshot := s.monitor.Snapshot()

	var workspaces []*dukhpb.WorkspaceStatus
	for _, wh := range snapshot {
		if req.WorkspaceName != "" && wh.Name != req.WorkspaceName {
			continue
		}
		ws := &dukhpb.WorkspaceStatus{
			Name: wh.Name,
			Path: wh.Path,
		}
		for _, ph := range wh.Projects {
			proj := &dukhpb.ProjectStatus{Name: ph.Name}
			for _, rh := range ph.Repositories {
				proj.Repositories = append(proj.Repositories, &dukhpb.RepoStatus{
					Name:             rh.Name,
					Path:             rh.Path,
					ConfiguredBranch: rh.ConfiguredBranch,
					ActualBranch:     rh.ActualBranch,
					Exists:           rh.Exists,
					BranchAligned:    rh.BranchAligned,
				})
			}
			ws.Projects = append(ws.Projects, proj)
		}
		workspaces = append(workspaces, ws)
	}

	return &dukhpb.StatusResponse{
		Workspaces:    workspaces,
		ServerVersion: version,
		UptimeSeconds: int64(time.Since(s.startedAt).Seconds()),
	}, nil
}

// WritePID writes the current process PID to $GRAZHDA_DIR/run/dukh.pid.
func WritePID(grazhdaDir string) error {
	runDir := filepath.Join(grazhdaDir, "run")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}
	pidFile := filepath.Join(runDir, "dukh.pid")
	return os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0o644)
}

// RemovePID removes the PID file on clean shutdown.
func RemovePID(grazhdaDir string) {
	_ = os.Remove(filepath.Join(grazhdaDir, "run", "dukh.pid"))
}
