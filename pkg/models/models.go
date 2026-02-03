package models

import "time"

type ScanResult struct {
	ImageRef    string                 `json:"image_ref"`
	ImageDigest string                 `json:"image_digest"`
	Platform    string                 `json:"platform"`
	StartedAt   time.Time              `json:"started_at"`
	FinishedAt  time.Time              `json:"finished_at"`
	Status      string                 `json:"status"` // PASS, WARN, FAIL
	Partial     bool                   `json:"partial"`
	Score       Score                  `json:"score"`
	Tools       map[string]*ToolResult `json:"tools"`
	Findings    Findings               `json:"findings"`
}

type Score struct {
	Total     int            `json:"total"`
	Grade     string         `json:"grade"` // A, B, C, D, F
	Breakdown map[string]int `json:"breakdown,omitempty"`
}

type ToolResult struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Status     string `json:"status"` // ok, failed, timeout, skipped
	Error      string `json:"error,omitempty"`
	OutputFile string `json:"output_file,omitempty"`
}

type Findings struct {
	Trivy      *TrivyFindings      `json:"trivy,omitempty"`
	Malcontent *MalcontentFindings `json:"malcontent,omitempty"`
	Magika     *MagikaFindings     `json:"magika,omitempty"`
}

type TrivyFindings struct {
	Vulnerabilities VulnCounts `json:"vulnerabilities"`
	Secrets         int        `json:"secrets"`
	Misconfigs      int        `json:"misconfigs"`
}

type VulnCounts struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

type MalcontentFindings struct {
	Total        int      `json:"total"`
	HighRisk     int      `json:"high_risk"`
	MediumRisk   int      `json:"medium_risk"`
	LowRisk      int      `json:"low_risk"`
	Capabilities []string `json:"capabilities,omitempty"`
}

type MagikaFindings struct {
	TotalFiles          int `json:"total_files"`
	SuspiciousTypes     int `json:"suspicious_types"`
	MismatchedExtension int `json:"mismatched_extension"`
}

// Config represents the scan configuration
type Config struct {
	ScanConfig ScanConfig `yaml:"scan_config"`
	Images     []Image    `yaml:"images"`
}

type ScanConfig struct {
	Schedule      string             `yaml:"schedule"`
	Platform      string             `yaml:"platform"`
	Notifications NotificationConfig `yaml:"notifications"`
}

type NotificationConfig struct {
	SlackWebhook  string `yaml:"slack_webhook"`
	OnFailure     bool   `yaml:"on_failure"`
	OnNewCritical bool   `yaml:"on_new_critical"`
}

type Image struct {
	Name           string   `yaml:"name"`
	Image          string   `yaml:"image"`
	Tags           []string `yaml:"tags"`
	Enabled        bool     `yaml:"enabled"`
	Private        bool     `yaml:"private"`
	RegistrySecret string   `yaml:"registry_secret,omitempty"`
}
