package server

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/vhula/grazhda/internal/config"
)

// defaultPeriod is used when config.dukh.monitoring.period_mins is zero or missing.
const defaultPeriod = 5 * time.Minute

// RepoHealth holds the last known health state of a single repository.
type RepoHealth struct {
	Name             string
	Path             string
	ConfiguredBranch string
	ActualBranch     string
	Exists           bool
	BranchAligned    bool
}

// ProjectHealth holds the health state of a project's repositories.
type ProjectHealth struct {
	Name         string
	Repositories []RepoHealth
}

// WorkspaceHealth holds the health state of all projects in a workspace.
type WorkspaceHealth struct {
	Name     string
	Path     string
	Projects []ProjectHealth
}

// Monitor polls workspace health and maintains an in-memory snapshot.
// The polling period is read from config on every cycle so changes to
// dukh.monitoring.period_mins take effect at the next tick without restart.
type Monitor struct {
	mu          sync.RWMutex
	snapshot    []WorkspaceHealth
	configPath  string
	logger      *log.Logger
	stopCh      chan struct{}
	doneCh      chan struct{}
	// triggerScan carries optional reply channels. A nil reply means
	// fire-and-forget (used by TriggerScan). A non-nil reply channel is
	// closed by the loop after the scan completes (used by TriggerScanAndWait).
	triggerScan chan chan struct{}
}

// NewMonitor creates a Monitor that reads workspace health from configPath.
func NewMonitor(configPath string, logger *log.Logger) *Monitor {
	return &Monitor{
		configPath:  configPath,
		logger:      logger,
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
		triggerScan: make(chan chan struct{}, 1),
	}
}

// Start begins the polling loop in a background goroutine.
func (m *Monitor) Start() {
	go m.loop()
}

// Stop signals the polling loop to exit and blocks until it has finished.
func (m *Monitor) Stop() {
	close(m.stopCh)
	<-m.doneCh
}

// TriggerScan sends a non-blocking fire-and-forget signal to perform an
// immediate rescan. If a scan is already queued the signal is dropped.
func (m *Monitor) TriggerScan() {
	select {
	case m.triggerScan <- nil: // nil reply = fire-and-forget
	default:
	}
}

// TriggerScanAndWait triggers an immediate rescan and blocks until it completes
// or ctx is cancelled. Returns ctx.Err() on cancellation.
func (m *Monitor) TriggerScanAndWait(ctx context.Context) error {
	reply := make(chan struct{})
	// Attempt to queue the request; wait if one is already in-flight.
	select {
	case m.triggerScan <- reply:
	case <-ctx.Done():
		return ctx.Err()
	}
	// Wait for the loop to close the reply channel after the scan finishes.
	select {
	case <-reply:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Snapshot returns a copy of the most recent workspace health data.
func (m *Monitor) Snapshot() []WorkspaceHealth {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cp := make([]WorkspaceHealth, len(m.snapshot))
	copy(cp, m.snapshot)
	return cp
}

// loop is the main polling goroutine. It uses time.Timer (not time.Ticker) so
// the period can be re-read from config after every scan.
func (m *Monitor) loop() {
	defer close(m.doneCh)

	m.scan()

	for {
		period := m.loadPeriod()
		timer := time.NewTimer(period)

		select {
		case <-timer.C:
			m.scan()

		case reply := <-m.triggerScan:
			timer.Stop()
			m.logger.Info("monitor: manual scan triggered")
			m.scan()
			if reply != nil {
				close(reply) // unblock any TriggerScanAndWait caller
			}

		case <-m.stopCh:
			timer.Stop()
			m.logger.Info("monitor: stopping")
			return
		}
	}
}

// loadPeriod reads the monitoring period from config, falling back to defaultPeriod.
func (m *Monitor) loadPeriod() time.Duration {
	cfg, err := config.Load(m.configPath)
	if err != nil || cfg.Dukh.Monitoring.PeriodMins <= 0 {
		return defaultPeriod
	}
	return time.Duration(cfg.Dukh.Monitoring.PeriodMins) * time.Minute
}

func (m *Monitor) scan() {
	cfg, err := config.Load(m.configPath)
	if err != nil {
		m.logger.Error("monitor: failed to load config", "err", err)
		return
	}

	result := make([]WorkspaceHealth, 0, len(cfg.Workspaces))
	for _, ws := range cfg.Workspaces {
		wh := WorkspaceHealth{Name: ws.Name, Path: ws.Path}
		for _, proj := range ws.Projects {
			ph := ProjectHealth{Name: proj.Name}
			for _, repo := range proj.Repositories {
				rh := m.checkRepo(ws.Path, proj, repo)
				ph.Repositories = append(ph.Repositories, rh)
			}
			wh.Projects = append(wh.Projects, ph)
		}
		result = append(result, wh)
	}

	m.mu.Lock()
	m.snapshot = result
	m.mu.Unlock()
	m.logger.Info("monitor: scan complete", "workspaces", len(result))
}

// checkRepo inspects a single repository on disk.
func (m *Monitor) checkRepo(wsPath string, proj config.Project, repo config.Repository) RepoHealth {
	destName := repo.LocalDirName
	if destName == "" {
		destName = repo.Name
	}
	repoPath := filepath.Join(wsPath, proj.Name, destName)

	configuredBranch := repo.Branch
	if configuredBranch == "" {
		configuredBranch = proj.Branch
	}

	rh := RepoHealth{
		Name:             repo.Name,
		Path:             repoPath,
		ConfiguredBranch: configuredBranch,
	}

	actualBranch, err := currentBranch(repoPath)
	if err != nil {
		rh.Exists = false
		rh.BranchAligned = false
		return rh
	}

	rh.Exists = true
	rh.ActualBranch = actualBranch
	rh.BranchAligned = actualBranch == configuredBranch
	return rh
}

// currentBranch returns the short HEAD branch name for the git repo at path.
func currentBranch(repoPath string) (string, error) {
	out, err := exec.Command("git", "-C", repoPath, "symbolic-ref", "--short", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
