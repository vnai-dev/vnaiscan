# @vnai-dev/vnaiscan

AI Agent Image Security Scanner - Scan container images for CVEs, secrets, malware, and suspicious binaries.

## Quick Start

```bash
# Run without installation
npx @vnai-dev/vnaiscan scan alpine:latest

# Or install globally
npm install -g @vnai-dev/vnaiscan
vnaiscan scan alpine:latest
```

## Requirements

- **Docker** - Required for running scans (the npm package uses Docker under the hood)
- **Docker socket access** - For scanning local images

## How it works

This npm package is a thin wrapper that runs vnaiscan via Docker. This ensures all security tools (Trivy, Malcontent, Magika) are available without manual installation.

```
npx @vnai-dev/vnaiscan scan image:tag
         ↓
docker run ghcr.io/vnai-dev/vnaiscan scan image:tag
```

## Usage

```bash
# Basic scan
npx @vnai-dev/vnaiscan scan myapp:latest

# Scan with JSON output
npx @vnai-dev/vnaiscan scan --output json myapp:latest

# Check tools
npx @vnai-dev/vnaiscan tools

# Get help
npx @vnai-dev/vnaiscan --help
```

## Alternative Installation Methods

For native binary installation (faster, no Docker required for CLI):

```bash
# Install script
curl -sSL https://scan.vnai.dev/install.sh | sh

# Homebrew
brew install vnai-dev/tap/vnaiscan
```

## Links

- [GitHub](https://github.com/vnai-dev/vnaiscan)
- [Documentation](https://scan.vnai.dev)
- [Docker Hub](https://ghcr.io/vnai-dev/vnaiscan)

## License

Apache-2.0
