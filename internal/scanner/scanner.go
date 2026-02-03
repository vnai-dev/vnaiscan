package scanner

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/vnai-dev/vnaiscan/internal/extractor"
	"github.com/vnai-dev/vnaiscan/pkg/models"
	"golang.org/x/sync/errgroup"
)

type ScanOptions struct {
	ImageRef       string
	Platform       string
	OutputFormat   string
	ReportDir      string
	TimeoutMins    int
	SkipTrivy      bool
	SkipMalcontent bool
	SkipMagika     bool
	Quiet          bool
	Verbose        bool
}

type Scanner struct {
	opts      ScanOptions
	workDir   string
	rootfsDir string
	results   *models.ScanResult
	mu        sync.Mutex
}

func New(opts ScanOptions) *Scanner {
	if opts.ReportDir == "" {
		opts.ReportDir = "./vnaiscan-reports"
	}
	return &Scanner{
		opts: opts,
		results: &models.ScanResult{
			ImageRef:  opts.ImageRef,
			Platform:  opts.Platform,
			StartedAt: time.Now(),
			Tools:     make(map[string]*models.ToolResult),
		},
	}
}

func (s *Scanner) Run(ctx context.Context) error {
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)

	if !s.opts.Quiet {
		bold.Printf("\n🔍 Scanning %s", s.opts.ImageRef)
		cyan.Printf(" (%s)\n\n", s.opts.Platform)
	}

	// Create work directory
	var err error
	s.workDir, err = os.MkdirTemp("", "vnaiscan-*")
	if err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}
	defer os.RemoveAll(s.workDir)

	s.rootfsDir = filepath.Join(s.workDir, "rootfs")
	if err := os.MkdirAll(s.rootfsDir, 0755); err != nil {
		return fmt.Errorf("failed to create rootfs directory: %w", err)
	}

	// Step 1: Pull and extract image
	if !s.opts.Quiet {
		fmt.Print("📦 Pulling image... ")
	}
	if err := s.pullAndExtract(ctx); err != nil {
		if !s.opts.Quiet {
			color.Red("✗\n")
		}
		return fmt.Errorf("failed to pull/extract image: %w", err)
	}
	if !s.opts.Quiet {
		green.Println("✓")
	}

	// Step 2: Run scanners in parallel
	if !s.opts.Quiet {
		fmt.Println("\n🛡️  Running security scanners...")
	}

	g, gctx := errgroup.WithContext(ctx)

	if !s.opts.SkipTrivy {
		g.Go(func() error {
			return s.runTrivy(gctx)
		})
	}

	if !s.opts.SkipMalcontent {
		g.Go(func() error {
			return s.runMalcontent(gctx)
		})
	}

	if !s.opts.SkipMagika {
		g.Go(func() error {
			return s.runMagika(gctx)
		})
	}

	// Wait for all scanners (don't fail on individual tool errors)
	_ = g.Wait()

	s.results.FinishedAt = time.Now()

	// Step 3: Calculate score and status
	s.calculateScore()

	// Step 4: Generate report
	if err := s.generateReport(); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Step 5: Print summary
	if !s.opts.Quiet {
		s.printSummary()
	}

	// Return error if status is FAIL
	if s.results.Status == "FAIL" {
		return fmt.Errorf("scan completed with FAIL status (score: %d)", s.results.Score.Total)
	}

	return nil
}

func (s *Scanner) pullAndExtract(ctx context.Context) error {
	// Pull image with platform
	pullCmd := exec.CommandContext(ctx, "docker", "pull", "--platform", s.opts.Platform, s.opts.ImageRef)
	if s.opts.Verbose {
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
	}
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("docker pull failed: %w", err)
	}

	// Get image digest
	inspectCmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.Id}}", s.opts.ImageRef)
	digestOut, err := inspectCmd.Output()
	if err != nil {
		return fmt.Errorf("docker inspect failed: %w", err)
	}
	s.results.ImageDigest = strings.TrimSpace(string(digestOut))

	// Create container (never run it - security)
	createCmd := exec.CommandContext(ctx, "docker", "create", "--platform", s.opts.Platform, s.opts.ImageRef)
	containerOut, err := createCmd.Output()
	if err != nil {
		return fmt.Errorf("docker create failed: %w", err)
	}
	containerID := strings.TrimSpace(string(containerOut))
	defer exec.Command("docker", "rm", containerID).Run()

	// Export and extract safely using our secure extractor
	// This prevents: symlinks, hardlinks, path traversal, device nodes
	exportCmd := exec.CommandContext(ctx, "docker", "export", containerID)
	pipe, err := exportCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}

	if err := exportCmd.Start(); err != nil {
		return fmt.Errorf("docker export failed to start: %w", err)
	}

	// Use safe extraction (skips symlinks, hardlinks, devices, validates paths)
	if err := extractor.SafeExtract(pipe, s.rootfsDir); err != nil {
		exportCmd.Wait()
		return fmt.Errorf("safe extraction failed: %w", err)
	}

	if err := exportCmd.Wait(); err != nil {
		return fmt.Errorf("docker export failed: %w", err)
	}

	// Make extracted files readable for scanning tools
	_ = makeReadableForScan(s.rootfsDir)

	return nil
}

func (s *Scanner) runTrivy(ctx context.Context) error {
	if !s.opts.Quiet {
		fmt.Println("  → Trivy (CVE/Secrets/Misconfig)...")
	}

	timeout := time.Duration(s.opts.TimeoutMins) * time.Minute
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	outputFile := filepath.Join(s.workDir, "trivy.json")
	cmd := exec.CommandContext(ctx, "trivy", "fs",
		"--format", "json",
		"--output", outputFile,
		"--scanners", "vuln,secret,misconfig",
		s.rootfsDir,
	)

	output, err := cmd.CombinedOutput()
	
	result := &models.ToolResult{
		Name:    "trivy",
		Version: getTrivyVersion(),
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.Status = "timeout"
		result.Error = "scan timed out"
	} else if err != nil {
		result.Status = "failed"
		result.Error = string(output)
	} else {
		result.Status = "ok"
		result.OutputFile = outputFile
		// Parse results would go here
	}

	s.mu.Lock()
	s.results.Tools["trivy"] = result
	s.mu.Unlock()

	return nil
}

func (s *Scanner) runMalcontent(ctx context.Context) error {
	if !s.opts.Quiet {
		fmt.Println("  → Malcontent (Binary Capabilities)...")
	}

	timeout := time.Duration(s.opts.TimeoutMins) * time.Minute
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	outputFile := filepath.Join(s.workDir, "malcontent.json")
	cmd := exec.CommandContext(ctx, "mal", "analyze",
		"--format", "json",
		"--output", outputFile,
		s.rootfsDir,
	)

	output, err := cmd.CombinedOutput()

	result := &models.ToolResult{
		Name:    "malcontent",
		Version: getMalcontentVersion(),
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.Status = "timeout"
		result.Error = "scan timed out"
	} else if err != nil {
		result.Status = "failed"
		result.Error = string(output)
	} else {
		result.Status = "ok"
		result.OutputFile = outputFile
	}

	s.mu.Lock()
	s.results.Tools["malcontent"] = result
	s.mu.Unlock()

	return nil
}

func (s *Scanner) runMagika(ctx context.Context) error {
	if !s.opts.Quiet {
		fmt.Println("  → Magika (AI File Type Detection)...")
	}

	timeout := time.Duration(s.opts.TimeoutMins) * time.Minute
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	outputFile := filepath.Join(s.workDir, "magika.json")
	cmd := exec.CommandContext(ctx, "magika",
		"--json",
		"--recursive",
		"--no-dereference",
		s.rootfsDir,
	)

	// Capture stdout to file, stderr for errors
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create magika output file: %w", err)
	}
	defer outFile.Close()
	cmd.Stdout = outFile

	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	err = cmd.Run()

	result := &models.ToolResult{
		Name:    "magika",
		Version: getMagikaVersion(),
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.Status = "timeout"
		result.Error = "scan timed out"
	} else if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("%v\n%s", err, stderrBuf.String())
	} else {
		result.Status = "ok"
		result.OutputFile = outputFile
	}

	s.mu.Lock()
	s.results.Tools["magika"] = result
	s.mu.Unlock()

	return nil
}

func (s *Scanner) calculateScore() {
	score := 0
	
	// TODO: Parse actual results and calculate score based on:
	// - Trivy: Critical=10, High=5, Medium=2, Low=1, Secrets=20
	// - Malcontent: High=15, Medium=7, Low=3
	// - Magika: Suspicious=5, Mismatched=3

	s.results.Score = models.Score{
		Total: score,
		Grade: calculateGrade(score),
	}

	// Check for tool failures first
	toolsOk := 0
	toolsFailed := 0
	for _, tool := range s.results.Tools {
		if tool.Status == "ok" {
			toolsOk++
		} else if tool.Status == "failed" || tool.Status == "timeout" {
			toolsFailed++
			s.results.Partial = true
		}
	}

	// Determine overall status
	enabledTools := len(s.results.Tools)
	
	if enabledTools > 0 && toolsOk == 0 {
		// All tools failed - this is an error state
		s.results.Status = "ERROR"
	} else if score > 100 {
		s.results.Status = "FAIL"
	} else if score > 30 || s.results.Partial {
		s.results.Status = "WARN"
	} else {
		s.results.Status = "PASS"
	}
}

// ExitCode returns appropriate exit code based on scan results
func (s *Scanner) ExitCode() int {
	switch s.results.Status {
	case "PASS":
		return 0
	case "WARN":
		return 1
	case "FAIL":
		return 2
	case "ERROR":
		return 3
	default:
		return 1
	}
}

func calculateGrade(score int) string {
	switch {
	case score <= 10:
		return "A"
	case score <= 30:
		return "B"
	case score <= 60:
		return "C"
	case score <= 100:
		return "D"
	default:
		return "F"
	}
}

func (s *Scanner) generateReport() error {
	if err := os.MkdirAll(s.opts.ReportDir, 0755); err != nil {
		return err
	}

	// Generate summary.json
	// TODO: Implement JSON/HTML/SARIF report generation

	return nil
}

func (s *Scanner) printSummary() {
	bold := color.New(color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	fmt.Println("\n" + strings.Repeat("━", 60))
	bold.Printf("📊 SCAN RESULTS")
	fmt.Printf("                                    Score: %d/100\n", s.results.Score.Total)
	fmt.Println(strings.Repeat("━", 60))

	for name, result := range s.results.Tools {
		var status string
		switch result.Status {
		case "ok":
			status = green.Sprint("✓ Passed")
		case "failed":
			status = red.Sprint("✗ Failed")
		case "timeout":
			status = yellow.Sprint("⏱ Timeout")
		default:
			status = yellow.Sprint("? Unknown")
		}
		fmt.Printf("\n🔬 %-20s %s\n", strings.Title(name), status)
	}

	fmt.Println("\n" + strings.Repeat("━", 60))

	var statusColor *color.Color
	switch s.results.Status {
	case "PASS":
		statusColor = green
	case "WARN":
		statusColor = yellow
	case "FAIL":
		statusColor = red
	default:
		statusColor = yellow
	}

	fmt.Printf("📋 Grade: %s (%s)", s.results.Score.Grade, statusColor.Sprint(s.results.Status))
	if s.results.Partial {
		yellow.Print(" [PARTIAL]")
	}
	fmt.Printf("   │   Report: %s\n", s.opts.ReportDir)
	fmt.Println(strings.Repeat("━", 60))
}

func getTrivyVersion() string {
	out, err := exec.Command("trivy", "--version").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func getMalcontentVersion() string {
	out, err := exec.Command("mal", "--version").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func getMagikaVersion() string {
	out, err := exec.Command("magika", "--version").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func CheckTools() error {
	tools := []struct {
		name    string
		cmd     string
		args    []string
		getVer  func() string
	}{
		{"Trivy", "trivy", []string{"--version"}, getTrivyVersion},
		{"Malcontent", "mal", []string{"--version"}, getMalcontentVersion},
		{"Magika", "magika", []string{"--version"}, getMagikaVersion},
		{"Docker", "docker", []string{"--version"}, func() string {
			out, _ := exec.Command("docker", "--version").Output()
			return strings.TrimSpace(string(out))
		}},
	}

	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	fmt.Print("\n🔧 Checking bundled tools...\n\n")

	allOk := true
	for _, tool := range tools {
		_, err := exec.LookPath(tool.cmd)
		if err != nil {
			red.Printf("  ✗ %s: not found\n", tool.name)
			allOk = false
		} else {
			ver := tool.getVer()
			green.Printf("  ✓ %s: %s\n", tool.name, ver)
		}
	}

	fmt.Println()
	if !allOk {
		return fmt.Errorf("some tools are missing")
	}
	return nil
}

// makeReadableForScan ensures extracted files are readable for scanning tools.
// Container exports often have restrictive permissions (e.g., /root 0700).
func makeReadableForScan(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // Skip inaccessible paths
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		mode := info.Mode()
		if d.IsDir() {
			// Ensure directory is traversable: u+rx
			perm := mode.Perm()
			newPerm := perm | 0500
			if newPerm != perm {
				_ = os.Chmod(path, (mode&^os.ModePerm)|newPerm)
			}
			return nil
		}

		if mode.IsRegular() {
			// Ensure file is readable: u+r
			perm := mode.Perm()
			newPerm := perm | 0400
			if newPerm != perm {
				_ = os.Chmod(path, (mode&^os.ModePerm)|newPerm)
			}
		}
		return nil
	})
}
