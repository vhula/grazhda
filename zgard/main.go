package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/vhula/grazhda/dukh/proto"
	"github.com/vhula/grazhda/internal/config"
)

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(args) < 2 {
		fmt.Println("Zgard - The Command CLI")
		fmt.Println("Usage: zgard <command>")
		fmt.Println("Commands:")
		fmt.Println("  ws init    - Initialize workspace")
		fmt.Println("  ws purge   - Purge workspace")
		fmt.Println("  run        - Run a custom script")
		return nil
	}

	command := args[1]
	switch command {
	case "ws":
		if len(args) < 3 {
			fmt.Println("Usage: zgard ws <subcommand>")
			return nil
		}
		subcommand := args[2]
		switch subcommand {
		case "init":
			log.Info("Initializing all workspaces...")

			conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.Dukh.Host, cfg.Dukh.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return fmt.Errorf("did not connect: %w", err)
			}
			defer conn.Close()

			client := proto.NewWorkspaceServiceClient(conn)
			resp, err := client.InitWorkspaces(context.Background(), &proto.InitWorkspacesRequest{})
			if err != nil {
				return fmt.Errorf("could not init workspace: %w", err)
			}
			for _, status := range resp.Statuses {
				log.Info(status)
			}
		case "purge":
			log.Info("Purging all workspaces...")

			conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.Dukh.Host, cfg.Dukh.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return fmt.Errorf("did not connect: %w", err)
			}
			defer conn.Close()

			client := proto.NewWorkspaceServiceClient(conn)
			resp, err := client.PurgeWorkspaces(context.Background(), &proto.PurgeWorkspacesRequest{})
			if err != nil {
				return fmt.Errorf("could not purge workspace: %w", err)
			}
			for _, status := range resp.Statuses {
				log.Info(status)
			}
		default:
			return fmt.Errorf("unknown ws subcommand: %s", subcommand)
		}
	case "run":
		log.Info("Running custom script...")
		// TODO: Implement script running
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
	return nil
}
