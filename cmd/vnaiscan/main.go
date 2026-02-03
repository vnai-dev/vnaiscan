package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vnai-dev/vnaiscan/internal/scanner"
)

var (
	version = "0.1.0"
	commit  = "dev"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "vnaiscan",
	Short: "AI Agent Image Security Scanner",
	Long: `vnaiscan - Security scanner for AI agent container images

Scans container images using multiple security tools:
  • Trivy     - CVE, secrets, and misconfiguration detection
  • Malcontent - Binary capability and malware behavior analysis
  • Magika    - AI-powered file type detection

Designed for scanning AI agents like OpenClaw, Moltbot, and similar projects.

Examples:
  vnaiscan scan ghcr.io/openclaw/openclaw:latest
  vnaiscan scan --platform linux/arm64 myimage:tag
  vnaiscan scan --output json --report ./reports image:tag`,
	Version: fmt.Sprintf("%s (commit: %s)", version, commit),
}

var scanCmd = &cobra.Command{
	Use:   "scan [image]",
	Short: "Scan a container image for security issues",
	Long: `Scan a container image using Trivy, Malcontent, and Magika.

The scan will:
  1. Pull the image (if not cached)
  2. Extract the filesystem
  3. Run all scanners in parallel
  4. Aggregate results and generate a unified report`,
	Args: cobra.ExactArgs(1),
	RunE: runScan,
}

var (
	flagPlatform   string
	flagOutput     string
	flagReportDir  string
	flagTimeout    int
	flagSkipTrivy  bool
	flagSkipMalc   bool
	flagSkipMagika bool
	flagQuiet      bool
	flagVerbose    bool
)

func init() {
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(toolsCmd)

	scanCmd.Flags().StringVarP(&flagPlatform, "platform", "p", "linux/amd64", "Target platform (e.g., linux/amd64, linux/arm64)")
	scanCmd.Flags().StringVarP(&flagOutput, "output", "o", "table", "Output format: table, json, sarif, html")
	scanCmd.Flags().StringVarP(&flagReportDir, "report", "r", "", "Directory to save reports (default: ./vnaiscan-reports)")
	scanCmd.Flags().IntVarP(&flagTimeout, "timeout", "t", 30, "Timeout per tool in minutes")
	scanCmd.Flags().BoolVar(&flagSkipTrivy, "skip-trivy", false, "Skip Trivy scanner")
	scanCmd.Flags().BoolVar(&flagSkipMalc, "skip-malcontent", false, "Skip Malcontent scanner")
	scanCmd.Flags().BoolVar(&flagSkipMagika, "skip-magika", false, "Skip Magika scanner")
	scanCmd.Flags().BoolVarP(&flagQuiet, "quiet", "q", false, "Minimal output")
	scanCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Verbose output")
}

func runScan(cmd *cobra.Command, args []string) error {
	imageRef := args[0]

	opts := scanner.ScanOptions{
		ImageRef:      imageRef,
		Platform:      flagPlatform,
		OutputFormat:  flagOutput,
		ReportDir:     flagReportDir,
		TimeoutMins:   flagTimeout,
		SkipTrivy:     flagSkipTrivy,
		SkipMalcontent: flagSkipMalc,
		SkipMagika:    flagSkipMagika,
		Quiet:         flagQuiet,
		Verbose:       flagVerbose,
	}

	s := scanner.New(opts)
	if err := s.Run(cmd.Context()); err != nil {
		return err
	}

	// Exit with appropriate code based on scan results
	// 0 = PASS, 1 = WARN, 2 = FAIL, 3 = ERROR
	exitCode := s.ExitCode()
	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("vnaiscan %s (commit: %s)\n", version, commit)
		fmt.Println("\nBundled tools:")
		fmt.Println("  Trivy:      https://trivy.dev")
		fmt.Println("  Malcontent: https://github.com/chainguard-dev/malcontent")
		fmt.Println("  Magika:     https://github.com/google/magika")
	},
}

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Check bundled tool versions and availability",
	RunE: func(cmd *cobra.Command, args []string) error {
		return scanner.CheckTools()
	},
}
