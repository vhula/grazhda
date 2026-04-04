package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/proto"
)

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	cmd := newRootCmd()
	cmd.SetArgs(args[1:])
	return cmd.Execute()
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "zgard",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Zgard - The Command CLI")
			fmt.Println("Usage: zgard <command>")
			fmt.Println("Commands:")
			fmt.Println("  ws init    - Initialize workspace")
			fmt.Println("  ws purge   - Purge workspace")
			fmt.Println("  dukh stop  - Stop Dukh server")
			fmt.Println("  dukh status - Show Dukh server status")
			fmt.Println("  run        - Run a custom script")
			return nil
		},
	}

	rootCmd.AddCommand(newWSCmd())
	rootCmd.AddCommand(newDukhCmd())
	rootCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Run a custom script",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("Running custom script...")
			return nil
		},
	})

	return rootCmd
}

func newWSCmd() *cobra.Command {
	wsCmd := &cobra.Command{
		Use:          "ws",
		Short:        "Manage workspaces",
		SilenceUsage: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown ws subcommand: %s", args[0])
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Usage: zgard ws <subcommand>")
			return nil
		},
	}

	wsCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			log.Info("Initializing all workspaces...")
			client, closeConn, err := newWorkspaceClient(cfg)
			if err != nil {
				return err
			}
			defer closeConn()

			resp, err := client.InitWorkspaces(context.Background(), &proto.InitWorkspacesRequest{})
			if err != nil {
				return fmt.Errorf("could not init workspace: %w", err)
			}
			for _, status := range resp.Statuses {
				log.Info(status)
			}
			return nil
		},
	})

	wsCmd.AddCommand(&cobra.Command{
		Use:   "purge",
		Short: "Purge workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			log.Info("Purging all workspaces...")
			client, closeConn, err := newWorkspaceClient(cfg)
			if err != nil {
				return err
			}
			defer closeConn()

			resp, err := client.PurgeWorkspaces(context.Background(), &proto.PurgeWorkspacesRequest{})
			if err != nil {
				return fmt.Errorf("could not purge workspace: %w", err)
			}
			for _, status := range resp.Statuses {
				log.Info(status)
			}
			return nil
		},
	})

	return wsCmd
}

func newDukhCmd() *cobra.Command {
	dukhCmd := &cobra.Command{
		Use:          "dukh",
		Short:        "Manage Dukh server",
		SilenceUsage: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown dukh subcommand: %s", args[0])
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Usage: zgard dukh <subcommand>")
			return nil
		},
	}

	dukhCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stop Dukh server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client, closeConn, err := newDukhClient(cfg)
			if err != nil {
				return err
			}
			defer closeConn()

			log.Info("Stopping Dukh server...")
			resp, err := client.StopDukh(context.Background(), &proto.StopDukhRequest{})
			if err != nil {
				return fmt.Errorf("could not stop dukh server: %w", err)
			}
			log.Info(resp.Status)
			return nil
		},
	})

	dukhCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show Dukh server status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client, closeConn, err := newDukhClient(cfg)
			if err != nil {
				return err
			}
			defer closeConn()

			resp, err := client.StatusDukh(context.Background(), &proto.StatusDukhRequest{})
			if err != nil {
				return fmt.Errorf("could not get dukh server status: %w", err)
			}
			if resp.Running {
				log.Info(resp.Status, "pid", resp.Pid)
				return nil
			}
			log.Info(resp.Status)
			return nil
		},
	})

	return dukhCmd
}

func newDukhClient(cfg *config.Config) (proto.DukhServiceClient, func(), error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.Dukh.Host, cfg.Dukh.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("did not connect: %w", err)
	}
	return proto.NewDukhServiceClient(conn), func() { _ = conn.Close() }, nil
}

func newWorkspaceClient(cfg *config.Config) (proto.WorkspaceServiceClient, func(), error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.Dukh.Host, cfg.Dukh.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("did not connect: %w", err)
	}
	return proto.NewWorkspaceServiceClient(conn), func() { _ = conn.Close() }, nil
}
