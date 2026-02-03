# vnaiscan

AI Agent Image Security Scanner - Scan container images for vulnerabilities, malicious behaviors, and suspicious files.

[![GitHub release](https://img.shields.io/github/v/release/vnai-dev/vnaiscan)](https://github.com/vnai-dev/vnaiscan/releases)
[![License](https://img.shields.io/github/license/vnai-dev/vnaiscan)](LICENSE)

## 🎯 What is vnaiscan?

**vnaiscan** is a security scanner specifically designed for AI agent container images like [OpenClaw](https://openclaw.ai), Moltbot, and similar projects. It combines multiple security tools to answer one question:

> **"Is this AI agent safe to run on my machine?"**

### Why vnaiscan?

Traditional container scanners focus on CVEs. But AI agents introduce unique risks:
- **Suspicious binary capabilities** (network access, file writes, process execution)
- **Hidden payloads** (malware disguised as images/models)
- **Supply chain attacks** (malicious dependencies)

vnaiscan uses three tools to cover all threat types:

| Tool | Version | Detects |
|------|---------|---------|
| **[Trivy](https://trivy.dev)** | 0.69.0 | CVEs, secrets, misconfigurations |
| **[Malcontent](https://github.com/chainguard-dev/malcontent)** | 1.15.1 | Binary capabilities, malware behaviors |
| **[Magika](https://github.com/google/magika)** | 1.0.1 | Disguised files, polyglots (AI-powered, rebuilt in Rust!) |

## 🚀 Quick Start

### Docker (Recommended)

No installation needed - everything bundled:

```bash
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/vnai-dev/vnaiscan scan alpine:latest
```

### npx (Node.js)

Run without installation - uses Docker under the hood:

```bash
npx @vnai-dev/vnaiscan scan alpine:latest
```

### Install Script

Downloads vnaiscan + all required tools:

```bash
# Full install (vnaiscan + trivy + malcontent + magika)
curl -sSL https://scan.vnai.dev/install.sh | sh

# Then run
vnaiscan scan alpine:latest
```

### Homebrew (macOS/Linux)

```bash
brew install vnai-dev/tap/vnaiscan
```

> Note: Homebrew installs vnaiscan + trivy. For malcontent/magika, use Docker mode or install manually.

## 📖 Usage

```bash
# Basic scan
vnaiscan scan ghcr.io/openclaw/openclaw:latest

# Scan with specific platform
vnaiscan scan --platform linux/arm64 myimage:tag

# Output as JSON
vnaiscan scan --output json image:tag

# Save reports to directory
vnaiscan scan --report ./my-reports image:tag

# Skip specific scanners
vnaiscan scan --skip-magika image:tag

# Check installed tools
vnaiscan tools
```

## 📊 Output Example

```
🔍 Scanning ghcr.io/openclaw/openclaw:v1.2.3 (linux/amd64)

📦 Pulling image... ✓

🛡️  Running security scanners...
  → Trivy (CVE/Secrets/Misconfig)...
  → Malcontent (Binary Capabilities)...
  → Magika (AI File Type Detection)...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 SCAN RESULTS                                    Score: 45/100
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔬 Trivy                 ✓ Passed
    Critical: 0  High: 2  Medium: 5  Low: 12

🔬 Malcontent            ⚠ Warning
    Suspicious capabilities: network_connect, file_write

🔬 Magika                ✓ Passed
    Scanned: 1,250 files  Suspicious: 0

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 Grade: C (WARN)   │   Report: ./vnaiscan-reports/
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## 📈 Risk Scoring

vnaiscan calculates a unified risk score:

| Finding | Points |
|---------|--------|
| Critical CVE | +10 |
| High CVE | +5 |
| Medium CVE | +2 |
| Low CVE | +1 |
| Secret detected | +20 |
| High-risk behavior | +15 |
| Medium-risk behavior | +7 |
| Suspicious file type | +5 |

| Grade | Score | Status |
|-------|-------|--------|
| A | 0-10 | PASS |
| B | 11-30 | PASS |
| C | 31-60 | WARN |
| D | 61-100 | WARN |
| F | 100+ | FAIL |

## 🔧 Configuration

Create `vnaiscan.yaml` in your project:

```yaml
# vnaiscan.yaml
platform: linux/amd64
timeout: 30  # minutes per tool

output:
  format: table  # table, json, sarif, html
  report_dir: ./vnaiscan-reports

tools:
  trivy:
    enabled: true
    severity: CRITICAL,HIGH,MEDIUM
  malcontent:
    enabled: true
  magika:
    enabled: true
    paths:
      - /usr/bin
      - /app
```

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md).

```bash
# Clone
git clone https://github.com/vnai-dev/vnaiscan.git
cd vnaiscan

# Build
go build -o vnaiscan ./cmd/vnaiscan

# Run tests
go test ./...

# Build Docker image
docker build -t vnaiscan .
```

## 📄 License

Apache License 2.0 - see [LICENSE](LICENSE)

## 🔗 Links

- **Website**: https://scan.vnai.dev
- **Documentation**: https://vnai.dev/docs/scan
- **GitHub**: https://github.com/vnai-dev/vnaiscan
- **Issues**: https://github.com/vnai-dev/vnaiscan/issues

---

Made with ❤️ by [vnai.dev](https://vnai.dev)
