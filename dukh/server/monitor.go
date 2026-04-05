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

// pollInterval is how often the monitor re-scans all workspaces.
const pollInterval = 30 * time.Second

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
type Monitor struct {
	mu         sync.RWMutex
	snapshot   []WorkspaceHealth
	configPath string
	logger     *log.Logger
	stop       chan struct{}
	done       chan struct{}
}

// NewMonitor creates a Monitor that will read from configPath.
func NewMonitor(configPath string, logger *log.Logger) *Monitor {
	return &Monitor{
		configPath: configPath,
		logger:     logger,
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
	}
}

// Start begins the polling loop in a background goroutine.
func (m *Monitor) Start() {
	go m.loop()
}

// Stop signals the polling loop to exit and waits for it to finish.
func (m *Monitor) Stop() {
	close(m.stop)
	<-m.done
}

// Snapshot returns a copy of the most recent workspace health data.
func (m *Monitor) Snapshot() []WorkspaceHealth {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cp := make([]WorkspaceHealth, len(m.snapshot))
	copy(cp, m.snapshot)
	return cp
}

func (m *Monitor) loop() {
	defer close(m.done)

	// Run an initial scan immediately, then poll on the interval.
	m.scan()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.scan()
		case <-m.stop:
			m.logger.Info("monitor: stopping")
			return
		}
	}
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
		// Directory doesn't exist or is not a git repo.
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

// resolveContext is a helper — wraps context to make the monitor stoppable.
func resolveContext(ctx context.Context) context.Context {
	return ctx
}
