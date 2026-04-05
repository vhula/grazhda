package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/ws"
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
			fmt.Println("  run        - Run a custom script")
			return nil
		},
	}

	rootCmd.AddCommand(newWSCmd())
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
			manager := ws.New(cfg)
			for _, status := range manager.Init() {
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
			manager := ws.New(cfg)
			for _, status := range manager.Purge() {
				log.Info(status)
			}
			return nil
		},
	})

	return wsCmd
}
