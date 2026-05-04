package workspace

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	fc "github.com/fatih/color"
	clr "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/executor"
)

// inspectRunOpts converts InspectOptions to RunOptions for use with shared targeting helpers.
func inspectRunOpts(opts InspectOptions) RunOptions {
	return RunOptions{
		ProjectName: opts.ProjectName,
		RepoName:    opts.RepoName,
		Tags:        opts.Tags,
		Parallel:    opts.Parallel,
		Verbose:     opts.Verbose,
	}
}

// ─────────────────────────────── Search ───────────────────────────────

// SearchMatch holds a single content-match or glob-match result.
type SearchMatch struct {
	Project string
	Repo    string
	File    string
	LineNo  int    // 0 for glob matches
	Content string // empty for glob matches
}

// Search fans out a pattern search across all resolved repositories.
// Content grep (default), filename glob (opts.Glob=true), or regex (opts.Regex=true).
func Search(ws config.Workspace, opts SearchOptions, out io.Writer) error {
	rOpts := inspectRunOpts(opts.InspectOptions)
	if err := ValidateFilters(ws, rOpts); err != nil {
		return err
	}
	if n := CountMatchingRepos(ws, rOpts); n > 1 {
		fmt.Fprintln(out, clr.Yellow(fmt.Sprintf(
			"Warning: --repo-name %q matches %d repositories", opts.RepoName, n)))
	}

	// compile regex once if needed
	var re *regexp.Regexp
	if opts.Regex && !opts.Glob {
		var err error
		re, err = regexp.Compile(opts.Pattern)
		if err != nil {
			return fmt.Errorf("invalid regex %q: %w", opts.Pattern, err)
		}
	}

	type searchJob struct {
		proj     config.Project
		repo     config.Repository
		repoPath string
	}

	wsPath := ExpandHome(ws.Path)
	var jobs []searchJob
	for _, proj := range ws.Projects {
		if opts.ProjectName != "" && proj.Name != opts.ProjectName {
			continue
		}
		projPath := filepath.Join(wsPath, proj.Name)
		for _, repo := range proj.Repositories {
			if !repoMatchesFilters(proj, repo, rOpts) {
				continue
			}
			dest := ResolveDestName(projPath, repo.Name, repo.LocalDirName, ResolveStructure(ws, proj))
			jobs = append(jobs, searchJob{proj, repo, filepath.Join(projPath, dest)})
		}
	}

	// results indexed by job position to preserve order
	results := make([][]SearchMatch, len(jobs))

	doJob := func(i int, j searchJob) {
		if _, err := os.Stat(j.repoPath); os.IsNotExist(err) {
			return
		}
		var matches []SearchMatch
		if opts.Glob {
			searchGlob(j.proj.Name, j.repo.Name, j.repoPath, opts.Pattern, &matches)
		} else {
			searchContent(j.proj.Name, j.repo.Name, j.repoPath, opts.Pattern, re, &matches)
		}
		results[i] = matches
	}

	if opts.Parallel {
		var wg sync.WaitGroup
		for i, j := range jobs {
			i, j := i, j
			wg.Add(1)
			go func() { defer wg.Done(); doJob(i, j) }()
		}
		wg.Wait()
	} else {
		for i, j := range jobs {
			doJob(i, j)
		}
	}

	totalMatches, reposWithMatches := 0, 0
	for _, matches := range results {
		if len(matches) == 0 {
			continue
		}
		reposWithMatches++
		totalMatches += len(matches)
		for _, m := range matches {
			if opts.Glob {
				fmt.Fprintf(out, "[%s/%s] %s\n", m.Project, m.Repo, m.File)
			} else {
				fmt.Fprintf(out, "[%s/%s] %s:%d: %s\n", m.Project, m.Repo, m.File, m.LineNo, m.Content)
			}
		}
	}
	fmt.Fprintf(out, "\n%d match(es) across %d repo(s)\n", totalMatches, reposWithMatches)
	return nil
}

func searchGlob(projName, repoName, repoPath, pattern string, out *[]SearchMatch) {
	_ = filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		if matched, _ := filepath.Match(pattern, d.Name()); matched {
			rel, _ := filepath.Rel(repoPath, path)
			*out = append(*out, SearchMatch{Project: projName, Repo: repoName, File: rel})
		}
		return nil
	})
}

func searchContent(projName, repoName, repoPath, pattern string, re *regexp.Regexp, out *[]SearchMatch) {
	_ = filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			scanFile(projName, repoName, repoPath, path, pattern, re, out)
		}
		return nil
	})
}

// scanFile scans a single file line-by-line.
// Binary files (null byte in first 512 bytes) are silently skipped.
func scanFile(projName, repoName, repoPath, filePath, pattern string, re *regexp.Regexp, out *[]SearchMatch) {
	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	header := make([]byte, 512)
	n, _ := f.Read(header)
	for _, b := range header[:n] {
		if b == 0 {
			return // binary file — skip
		}
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return
	}

	rel, _ := filepath.Rel(repoPath, filePath)
	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		var matched bool
		if re != nil {
			matched = re.MatchString(line)
		} else {
			matched = strings.Contains(line, pattern)
		}
		if matched {
			*out = append(*out, SearchMatch{
				Project: projName, Repo: repoName,
				File: rel, LineNo: lineNo, Content: line,
			})
		}
	}
}

// ─────────────────────────────── Diff ────────────────────────────────

type diffRow struct {
	repo        string
	uncommitted int // -1 = not cloned
	ahead       int // -2 = no upstream, -1 = not cloned
	behind      int // -2 = no upstream, -1 = not cloned
}

// Diff shows per-repo Git state (uncommitted, ahead, behind) in project-grouped aligned tables.
func Diff(ws config.Workspace, exec executor.Executor, opts InspectOptions, out io.Writer) error {
	rOpts := inspectRunOpts(opts)
	if err := ValidateFilters(ws, rOpts); err != nil {
		return err
	}
	if n := CountMatchingRepos(ws, rOpts); n > 1 {
		fmt.Fprintln(out, clr.Yellow(fmt.Sprintf(
			"Warning: --repo-name %q matches %d repositories", opts.RepoName, n)))
	}

	wsPath := ExpandHome(ws.Path)
	fmt.Fprintln(out, clr.Blue("Workspace: "+ws.Name))

	totalClean, totalDirty, totalNotCloned := 0, 0, 0

	for _, proj := range ws.Projects {
		if opts.ProjectName != "" && proj.Name != opts.ProjectName {
			continue
		}
		projPath := filepath.Join(wsPath, proj.Name)

		type diffJob struct {
			repo     config.Repository
			repoPath string
		}
		var jobs []diffJob
		for _, repo := range proj.Repositories {
			if !repoMatchesFilters(proj, repo, rOpts) {
				continue
			}
			dest := ResolveDestName(projPath, repo.Name, repo.LocalDirName, ResolveStructure(ws, proj))
			jobs = append(jobs, diffJob{repo, filepath.Join(projPath, dest)})
		}
		if len(jobs) == 0 {
			continue
		}

		rows := make([]diffRow, len(jobs))
		doJob := func(i int, j diffJob) {
			if _, err := os.Stat(j.repoPath); os.IsNotExist(err) {
				rows[i] = diffRow{
					repo: lastSegment(j.repo.Name), uncommitted: -1, ahead: -1, behind: -1,
				}
				return
			}
			rows[i] = collectDiffRow(j.repoPath, lastSegment(j.repo.Name), exec)
		}

		if opts.Parallel {
			var wg sync.WaitGroup
			for i, j := range jobs {
				i, j := i, j
				wg.Add(1)
				go func() { defer wg.Done(); doJob(i, j) }()
			}
			wg.Wait()
		} else {
			for i, j := range jobs {
				doJob(i, j)
			}
		}

		fmt.Fprintf(out, "  %s\n\n", clr.Blue("Project: "+proj.Name))

		headers := []string{"REPO", "UNCOMMITTED", "AHEAD", "BEHIND"}
		tableRows := make([][]string, len(rows))
		rowColors := make([]func(string) string, len(rows))
		for i, row := range rows {
			tableRows[i] = []string{
				row.repo,
				diffUncommittedStr(row.uncommitted),
				diffAheadBehindStr(row.ahead),
				diffAheadBehindStr(row.behind),
			}
			switch {
			case row.uncommitted == -1:
				rowColors[i] = func(s string) string { return clr.Yellow(s) }
				totalNotCloned++
			case row.uncommitted > 0:
				rowColors[i] = func(s string) string { return clr.Red(s) }
				totalDirty++
			case row.ahead > 0 || row.behind > 0:
				rowColors[i] = func(s string) string { return clr.Yellow(s) }
				totalDirty++
			default:
				rowColors[i] = func(s string) string { return clr.Green(s) }
				totalClean++
			}
		}
		printTable(out, "    ", headers, tableRows, rowColors)
		fmt.Fprintln(out)
	}

	fmt.Fprintf(out, "%s %d clean  %s %d dirty  %s %d not cloned\n",
		clr.Green("✓"), totalClean,
		clr.Red("✗"), totalDirty,
		clr.Yellow("⏭"), totalNotCloned,
	)
	return nil
}

func collectDiffRow(repoPath, displayName string, exec executor.Executor) diffRow {
	row := diffRow{repo: displayName}

	statusOut, _ := exec.RunCapture(repoPath, "git status --porcelain")
	row.uncommitted = len(nonEmptyLines(statusOut))

	aheadOut, err := exec.RunCapture(repoPath, "git rev-list @{u}..HEAD --count")
	if err != nil {
		row.ahead = -2 // no upstream
	} else {
		row.ahead, _ = strconv.Atoi(strings.TrimSpace(aheadOut))
	}

	behindOut, err := exec.RunCapture(repoPath, "git rev-list HEAD..@{u} --count")
	if err != nil {
		row.behind = -2 // no upstream
	} else {
		row.behind, _ = strconv.Atoi(strings.TrimSpace(behindOut))
	}

	return row
}

func diffUncommittedStr(n int) string {
	if n == -1 {
		return "(not cloned)"
	}
	return strconv.Itoa(n)
}

func diffAheadBehindStr(n int) string {
	switch n {
	case -1:
		return "(not cloned)"
	case -2:
		return "--"
	default:
		return strconv.Itoa(n)
	}
}

// ─────────────────────────────── Stats ───────────────────────────────

type statsRow struct {
	repo         string
	lastCommit   string // "YYYY-MM-DD HH:MM" or "(not cloned)"
	commits30d   int    // -1 = not cloned
	contributors int    // -1 = not cloned
}

// Stats shows per-repo metadata (last commit, 30-day commits, unique contributors) in project-grouped tables.
func Stats(ws config.Workspace, exec executor.Executor, opts InspectOptions, out io.Writer) error {
	rOpts := inspectRunOpts(opts)
	if err := ValidateFilters(ws, rOpts); err != nil {
		return err
	}
	if n := CountMatchingRepos(ws, rOpts); n > 1 {
		fmt.Fprintln(out, clr.Yellow(fmt.Sprintf(
			"Warning: --repo-name %q matches %d repositories", opts.RepoName, n)))
	}

	wsPath := ExpandHome(ws.Path)
	fmt.Fprintln(out, clr.Blue("Workspace: "+ws.Name))

	for _, proj := range ws.Projects {
		if opts.ProjectName != "" && proj.Name != opts.ProjectName {
			continue
		}
		projPath := filepath.Join(wsPath, proj.Name)

		type statsJob struct {
			repo     config.Repository
			repoPath string
		}
		var jobs []statsJob
		for _, repo := range proj.Repositories {
			if !repoMatchesFilters(proj, repo, rOpts) {
				continue
			}
			dest := ResolveDestName(projPath, repo.Name, repo.LocalDirName, ResolveStructure(ws, proj))
			jobs = append(jobs, statsJob{repo, filepath.Join(projPath, dest)})
		}
		if len(jobs) == 0 {
			continue
		}

		rows := make([]statsRow, len(jobs))
		doJob := func(i int, j statsJob) {
			if _, err := os.Stat(j.repoPath); os.IsNotExist(err) {
				rows[i] = statsRow{
					repo: lastSegment(j.repo.Name), lastCommit: "(not cloned)",
					commits30d: -1, contributors: -1,
				}
				return
			}
			rows[i] = collectStatsRow(j.repoPath, lastSegment(j.repo.Name), exec)
		}

		if opts.Parallel {
			var wg sync.WaitGroup
			for i, j := range jobs {
				i, j := i, j
				wg.Add(1)
				go func() { defer wg.Done(); doJob(i, j) }()
			}
			wg.Wait()
		} else {
			for i, j := range jobs {
				doJob(i, j)
			}
		}

		fmt.Fprintf(out, "  %s\n\n", clr.Blue("Project: "+proj.Name))

		headers := []string{"REPO", "LAST COMMIT", "30D COMMITS", "CONTRIBUTORS"}
		tableRows := make([][]string, len(rows))
		for i, row := range rows {
			c30 := strconv.Itoa(row.commits30d)
			if row.commits30d == -1 {
				c30 = "-"
			}
			ctb := strconv.Itoa(row.contributors)
			if row.contributors == -1 {
				ctb = "-"
			}
			tableRows[i] = []string{row.repo, row.lastCommit, c30, ctb}
		}
		printTable(out, "    ", headers, tableRows, nil)
		fmt.Fprintln(out)
	}
	return nil
}

func collectStatsRow(repoPath, displayName string, exec executor.Executor) statsRow {
	row := statsRow{repo: displayName}

	lastOut, err := exec.RunCapture(repoPath, `git log -1 --format="%ci"`)
	if err != nil || strings.TrimSpace(lastOut) == "" {
		row.lastCommit = "--"
	} else {
		row.lastCommit = parseCommitDate(strings.TrimSpace(lastOut))
	}

	countOut, err := exec.RunCapture(repoPath, `git log --since="30 days ago" --format="%H"`)
	if err != nil {
		row.commits30d = 0
	} else {
		row.commits30d = len(nonEmptyLines(countOut))
	}

	emailOut, err := exec.RunCapture(repoPath, `git log --format="%ae"`)
	if err != nil {
		row.contributors = 0
	} else {
		seen := map[string]struct{}{}
		for _, e := range nonEmptyLines(emailOut) {
			seen[strings.ToLower(strings.TrimSpace(e))] = struct{}{}
		}
		row.contributors = len(seen)
	}

	return row
}

func parseCommitDate(s string) string {
	// "2024-01-15 10:30:22 +0000" or ISO-8601 variants — take first 16 chars
	s = strings.ReplaceAll(s, "T", " ")
	if len(s) >= 16 {
		return s[:16]
	}
	return s
}

// ─────────────────────────────── Shared helpers ────────────────────────────────

// nonEmptyLines returns the non-empty lines of s.
func nonEmptyLines(s string) []string {
	var out []string
	for _, l := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

// printTable prints a column-aligned table with a bold header row, a styled
// separator, and data rows. indent is prepended to every line.
// rowColors may be nil; otherwise rowColors[i] (if non-nil) colourises the
// entire formatted data row string after padding is applied.
func printTable(out io.Writer, indent string, headers []string, rows [][]string, rowColors []func(string) string) {
	if len(rows) == 0 {
		return
	}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i := range headers {
			if i < len(row) && len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	headerStyle := lipgloss.NewStyle().Bold(true)
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// header row — bold unless colors are disabled
	var hdr strings.Builder
	hdr.WriteString(indent)
	for i, h := range headers {
		padded := fmt.Sprintf("%-*s", widths[i], h)
		if !fc.NoColor {
			hdr.WriteString(headerStyle.Render(padded))
		} else {
			hdr.WriteString(padded)
		}
		if i < len(headers)-1 {
			hdr.WriteString("  ")
		}
	}
	fmt.Fprintln(out, hdr.String())

	// separator
	total := 0
	for i, w := range widths {
		total += w
		if i < len(widths)-1 {
			total += 2
		}
	}
	sep := strings.Repeat("─", total)
	if !fc.NoColor {
		fmt.Fprintf(out, "%s%s\n", indent, sepStyle.Render(sep))
	} else {
		fmt.Fprintf(out, "%s%s\n", indent, sep)
	}

	// data rows
	for ri, row := range rows {
		var sb strings.Builder
		sb.WriteString(indent)
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			if i < len(headers)-1 {
				fmt.Fprintf(&sb, "%-*s  ", widths[i], cell)
			} else {
				sb.WriteString(cell)
			}
		}
		line := sb.String()
		if rowColors != nil && ri < len(rowColors) && rowColors[ri] != nil {
			line = rowColors[ri](line)
		}
		fmt.Fprintln(out, line)
	}
}
