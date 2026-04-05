package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/dukh/server"
)

func main() {
	rootCmd := &cobra.Command{
		Use:           "dukh",
		Short:         "Dukh — Grazhda workspace health monitor",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	rootCmd.AddCommand(startCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the dukh gRPC server",
		RunE:  runStart,
	}
}

func runStart(_ *cobra.Command, _ []string) error {
	grazhdaDir := os.Getenv("GRAZHDA_DIR")
	if grazhdaDir == "" {
		return fmt.Errorf("GRAZHDA_DIR environment variable is not set")
	}

	logger, cleanup, err := server.InitLogger(grazhdaDir)
	if err != nil {
		return fmt.Errorf("initialising logger: %w", err)
	}
	defer cleanup()

	configPath := filepath.Join(grazhdaDir, "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("config not found at %s: %w", configPath, err)
	}

	if err := server.WritePID(grazhdaDir); err != nil {
		logger.Warn("could not write PID file", "err", err)
	}
	defer server.RemovePID(grazhdaDir)

	monitor := server.NewMonitor(configPath, logger)
	monitor.Start()
	defer monitor.Stop()

	srv := server.New(monitor, logger)
	addr := "localhost:50501"
	logger.Info("dukh starting", "addr", addr, "config", configPath)

	if err := srv.ListenAndServe(addr); err != nil {
		// GracefulStop causes Serve to return a non-nil error — treat it as clean exit.
		logger.Info("dukh stopped", "reason", err)
	}
	return nil
}
