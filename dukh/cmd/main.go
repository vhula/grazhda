package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	pb "github.com/vhula/grazhda/dukh/proto"

	"google.golang.org/grpc"
)

type workspaceServer struct {
	pb.UnimplementedWorkspaceServiceServer
	workspaces map[string]*pb.Workspace
	mu         sync.RWMutex
}

func (s *workspaceServer) CreateWorkspace(ctx context.Context, req *pb.CreateWorkspaceRequest) (*pb.CreateWorkspaceResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("ws-%d", time.Now().Unix())
	ws := &pb.Workspace{
		Id:        id,
		Name:      req.Name,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	s.workspaces[id] = ws

	return &pb.CreateWorkspaceResponse{
		Id:     id,
		Status: "created",
	}, nil
}

func (s *workspaceServer) ListWorkspaces(ctx context.Context, req *pb.ListWorkspacesRequest) (*pb.ListWorkspacesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var workspaces []*pb.Workspace
	for _, ws := range s.workspaces {
		workspaces = append(workspaces, ws)
	}

	return &pb.ListWorkspacesResponse{
		Workspaces: workspaces,
	}, nil
}

func (s *workspaceServer) DeleteWorkspace(ctx context.Context, req *pb.DeleteWorkspaceRequest) (*pb.DeleteWorkspaceResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workspaces[req.Id]; !exists {
		return &pb.DeleteWorkspaceResponse{
			Status: "not found",
		}, nil
	}

	delete(s.workspaces, req.Id)

	return &pb.DeleteWorkspaceResponse{
		Status: "deleted",
	}, nil
}

func startServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := &workspaceServer{
		workspaces: make(map[string]*pb.Workspace),
	}

	s := grpc.NewServer()
	pb.RegisterWorkspaceServiceServer(s, server)

	log.Println("Dukh gRPC server starting on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Dukh - The Worker CLI")
		fmt.Println("Usage: dukh <command>")
		fmt.Println("Commands:")
		fmt.Println("  start    - Start the Dukh gRPC server")
		fmt.Println("  stop     - Stop the Dukh server")
		fmt.Println("  status   - Check server status")
		return
	}

	command := os.Args[1]
	switch command {
	case "start":
		fmt.Println("Starting Dukh gRPC server...")
		startServer()
	case "stop":
		fmt.Println("Stopping Dukh server...")
		// TODO: Implement server stop logic
	case "status":
		fmt.Println("Dukh server status: Not implemented yet")
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}
