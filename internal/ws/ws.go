package ws

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/vhula/grazhda/internal/config"
	internalexec "github.com/vhula/grazhda/internal/exec"
)

type Workspace struct {
	ID           string
	Name         string
	CreatedAt    string
	AbsolutePath string
}

type Workspaces struct {
	config *config.Config
	items  map[string]*Workspace
}

func New(cfg *config.Config) *Workspaces {
	return &Workspaces{
		config: cfg,
		items:  make(map[string]*Workspace),
	}
}

func (w *Workspaces) Init() []string {
	statuses := make([]string, 0)

	for _, wsConfig := range w.config.Workspaces {
		if err := os.MkdirAll(wsConfig.Path, 0755); err != nil {
			statuses = append(statuses, fmt.Sprintf("failed to create dir for %s: %v", wsConfig.Name, err))
			continue
		}

		createdAt := time.Now().Format(time.RFC3339)
		logPath := filepath.Join(wsConfig.Path, "dukh.log")
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			statuses = append(statuses, fmt.Sprintf("failed to create log for %s: %v", wsConfig.Name, err))
			continue
		}

		_, writeErr := fmt.Fprintf(file, "Id: %s, Name: %s, CreatedAt: %s\n", wsConfig.Name, wsConfig.Name, createdAt)
		closeErr := file.Close()
		if writeErr != nil {
			statuses = append(statuses, fmt.Sprintf("failed to write log for %s: %v", wsConfig.Name, writeErr))
			continue
		}
		if closeErr != nil {
			statuses = append(statuses, fmt.Sprintf("failed to close log for %s: %v", wsConfig.Name, closeErr))
			continue
		}

		w.items[wsConfig.Name] = &Workspace{
			ID:           wsConfig.Name,
			Name:         wsConfig.Name,
			CreatedAt:    createdAt,
			AbsolutePath: wsConfig.Path,
		}
		statuses = append(statuses, fmt.Sprintf("initialized workspace %s", wsConfig.Name))

		for _, project := range wsConfig.Projects {
			projectDir := filepath.Join(wsConfig.Path, project.Name)
			if err := os.MkdirAll(projectDir, 0755); err != nil {
				statuses = append(statuses, fmt.Sprintf("failed to create project dir for %s/%s: %v", wsConfig.Name, project.Name, err))
				continue
			}

			if len(project.Subprojects) > 0 {
				for _, subproject := range project.Subprojects {
					for _, repo := range subproject.Repositories {
						if err := cloneRepository(wsConfig.CloneCommandTemplate, repo.Name, repo.LocalDirName, subproject.Branch, projectDir); err != nil {
							statuses = append(statuses, fmt.Sprintf("failed to clone %s: %v", repo.Name, err))
							continue
						}
						statuses = append(statuses, fmt.Sprintf("cloned %s successfully", repo.Name))
					}
				}
				continue
			}

			for _, repo := range project.Repositories {
				if err := cloneRepository(wsConfig.CloneCommandTemplate, repo.Name, repo.LocalDirName, project.Branch, projectDir); err != nil {
					statuses = append(statuses, fmt.Sprintf("failed to clone %s: %v", repo.Name, err))
					continue
				}
				statuses = append(statuses, fmt.Sprintf("cloned %s successfully", repo.Name))
			}
		}
	}

	return statuses
}

func (w *Workspaces) Purge() []string {
	statuses := make([]string, 0, len(w.config.Workspaces))
	for _, wsConfig := range w.config.Workspaces {
		if err := os.RemoveAll(wsConfig.Path); err != nil {
			statuses = append(statuses, fmt.Sprintf("failed to delete directory for %s: %v", wsConfig.Name, err))
			continue
		}
		delete(w.items, wsConfig.Name)
		statuses = append(statuses, fmt.Sprintf("purged %s at %s", wsConfig.Name, wsConfig.Path))
	}
	return statuses
}

func cloneRepository(templateCommand, repoName, localDirName, branch, projectDir string) error {
	command, err := constructGitCommand(templateCommand, repoName, branch, localDirName)
	if err != nil {
		return err
	}
	output, err := internalexec.RunShell(command, projectDir)
	if err != nil {
		trimmed := strings.TrimSpace(output)
		if trimmed == "" {
			return err
		}
		return fmt.Errorf("%w: %s", err, trimmed)
	}
	return nil
}

func constructGitCommand(templateCommand, repoName, branch, localDirName string) (string, error) {
	destDir := localDirName
	if destDir == "" {
		destDir = repoName
	}

	tmpl, err := template.New("git-clone").Parse(templateCommand)
	if err != nil {
		return "", fmt.Errorf("failed to parse clone command template: %w", err)
	}

	data := struct {
		RepoName string
		Branch   string
		DestDir  string
	}{
		RepoName: repoName,
		Branch:   branch,
		DestDir:  destDir,
	}

	buf := &strings.Builder{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", fmt.Errorf("failed to execute clone command template: %w", err)
	}

	return buf.String(), nil
}
