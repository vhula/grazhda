package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	dukhpb "github.com/vhula/grazhda/dukh/proto"
	icolor "github.com/vhula/grazhda/internal/color"
)

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show dukh process status (running, PID, uptime)",
		RunE:  runDukhStatus,
	}
}

func runDukhStatus(_ *cobra.Command, _ []string) error {
	grazhdaDir := os.Getenv("GRAZHDA_DIR")
	if grazhdaDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %w", err)
		}
		grazhdaDir = filepath.Join(home, ".grazhda")
	}

	pid, err := readPIDFile(grazhdaDir)
	if err != nil {
		fmt.Println(icolor.Yellow("○") + "  dukh: not running")
		return nil
	}
	if pid == 0 {
		fmt.Println(icolor.Yellow("○") + "  dukh: not running")
		return nil
	}

	if !isProcessAlive(pid) {
		fmt.Printf("%s  dukh: not running  %s\n",
			icolor.Yellow("○"),
			icolor.Yellow(fmt.Sprintf("(stale pid %d — removing pid file)", pid)),
		)
		_ = os.Remove(filepath.Join(grazhdaDir, "run", "dukh.pid"))
		return nil
	}

	// Try gRPC with a short timeout to get uptime from the live server.
	uptime := tryGetUptime()
	if uptime != "" {
		fmt.Printf("%s  dukh: %s  (pid %d, uptime: %s)\n",
			icolor.Green("●"),
			icolor.Green("running"),
			pid,
			icolor.Blue(uptime),
		)
	} else {
		fmt.Printf("%s  dukh: %s  (pid %d)\n",
			icolor.Green("●"),
			icolor.Green("running"),
			pid,
		)
	}
	return nil
}

// readPIDFile reads and returns the PID stored in $GRAZHDA_DIR/run/dukh.pid.
// Returns 0 if the file does not exist or cannot be parsed.
func readPIDFile(grazhdaDir string) (int, error) {
	pidPath := filepath.Join(grazhdaDir, "run", "dukh.pid")
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

// tryGetUptime attempts a quick gRPC Status call to retrieve the server's
// uptime. Returns an empty string if dukh is unreachable or the call fails.
func tryGetUptime() string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	conn, client, err := dial()
	if err != nil {
		return ""
	}
	defer conn.Close()

	resp, err := client.Status(ctx, &dukhpb.StatusRequest{})
	if err != nil {
		return ""
	}
	return formatUptime(time.Duration(resp.UptimeSeconds) * time.Second)
}

// formatUptime renders a duration as a human-readable string.
func formatUptime(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
