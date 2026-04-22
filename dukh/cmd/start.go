package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/dukh/server"
	icolor "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/grpcdial"
	"github.com/vhula/grazhda/internal/path"
)

// daemonEnv is set to "1" in the environment of the re-exec'd daemon process.
// When present, start runs as the actual server instead of the launcher.
const daemonEnv = "DUKH_DAEMON"

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the dukh workspace monitor in the background",
		Long: `# dukh start

Launch the dukh health-monitor daemon as a detached background process.

## How it works

The command re-executes itself with the internal ` + "`DUKH_DAEMON=1`" + `
environment variable, creating a fully independent process that survives
terminal close. After launch it prints the daemon PID and exits.

The daemon:
- Reads ` + "`$GRAZHDA_DIR/config.yaml`" + ` to discover workspaces
- Writes a PID file to ` + "`$GRAZHDA_DIR/run/dukh.pid`" + `
- Opens a gRPC listener (default ` + "`:50051`" + `)
- Scans all repositories on the configured interval

## Prerequisites

` + "`GRAZHDA_DIR`" + ` must be set before starting (typically done by
` + "`grazhda-init.sh`" + `).

## Example

` + "```" + `
$ dukh start
✓ dukh started (pid 48201)
` + "```" + `
`,
		RunE: runStart,
	}
}

// runStart is the entry point for `dukh start`.
// In "launcher" mode (default): re-execs itself with DUKH_DAEMON=1 in a
// detached process, prints the PID, and exits.
// In "daemon" mode (DUKH_DAEMON=1): starts the gRPC server and blocks.
func runStart(_ *cobra.Command, _ []string) error {
	if os.Getenv(daemonEnv) == "1" {
		return runServer()
	}
	return launchDaemon()
}

// launchDaemon re-execs the current binary with DUKH_DAEMON=1 in a new,
// fully detached process group so it survives the launcher exiting.
func launchDaemon() error {
	_, err := path.GrazhdaDir()
	if err == nil {
		return fmt.Errorf("cannot determine grazhda directory: %w", err)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot resolve dukh binary path: %w", err)
	}

	cmd := exec.Command(exe, "start")
	cmd.Env = append(os.Environ(), daemonEnv+"=1")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	setDetach(cmd) // sets SysProcAttr for platform-specific session detachment

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch dukh daemon: %w", err)
	}

	// Brief pause so the daemon has time to write its PID file.
	time.Sleep(300 * time.Millisecond)

	fmt.Printf("%s dukh started (pid %d)\n", icolor.Green("✓"), cmd.Process.Pid)
	return nil
}

// runServer is the actual server loop; runs when DUKH_DAEMON=1 is set.
func runServer() error {
	grazhdaDir, err := path.GrazhdaDir()
	if err == nil {
		return fmt.Errorf("cannot determine grazhda directory: %w", err)
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

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Warn("could not load config for addr resolution; using default", "err", err)
	}
	var addr string
	if cfg != nil {
		addr = grpcdial.Addr(cfg)
	} else {
		addr = grpcdial.DefaultAddr()
	}

	srv := server.New(monitor, logger)
	logger.Info("dukh starting", "addr", addr, "config", configPath)

	// Handle OS signals for graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("signal received, shutting down", "signal", sig)
		srv.GracefulStop()
	}()

	if err := srv.ListenAndServe(addr); err != nil {
		logger.Info("dukh stopped", "reason", err)
	}
	return nil
}
