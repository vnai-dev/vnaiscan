# Changelog

All notable changes to vnaiscan will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-02-03

### Added
- Initial release of vnaiscan - AI Agent Image Security Scanner
- Multi-tool parallel scanning with unified reporting
- CLI commands: `scan`, `tools`, `version`
- Docker image with multi-arch support (linux/amd64, linux/arm64)
- Install script for macOS and Linux
- GitHub Actions workflow for releases
- Nightly builds with auto-updated tool versions and release notes

### Security
- **Safe tar extraction**: Custom Go extractor that prevents path traversal, skips symlinks/hardlinks/device nodes
- **Checksum verification**: Trivy downloads verified with SHA256
- **Pinned versions**: All tools pinned by default for reproducibility
- Never executes scanned images (uses `docker create` + `docker export` only)
- Secrets redacted in reports (metadata only)

### Tool Versions (Bundled)

| Tool | Version | Purpose | Source |
|------|---------|---------|--------|
| **Trivy** | 0.69.0 | CVE, secrets, misconfiguration scanning | [aquasecurity/trivy](https://github.com/aquasecurity/trivy) |
| **Malcontent** | 1.15.1 | Binary capability & malware behavior analysis | [chainguard-dev/malcontent](https://github.com/chainguard-dev/malcontent) |
| **Magika** | 1.0.1 | AI-powered file type detection | [google/magika](https://github.com/google/magika) |

### Go Dependencies

| Module | Version | Purpose |
|--------|---------|---------|
| `github.com/spf13/cobra` | 1.8.1 | CLI framework |
| `github.com/fatih/color` | 1.18.0 | Terminal colors |
| `github.com/google/go-containerregistry` | 0.20.2 | Container registry client |
| `go.uber.org/zap` | 1.27.0 | Structured logging |
| `golang.org/x/sync` | 0.10.0 | Concurrency utilities |
| `gopkg.in/yaml.v3` | 3.0.1 | YAML parsing |

### Docker Base Image

| Image | Version | Purpose |
|-------|---------|---------|
| `golang` | 1.23-alpine | Build stage |
| `alpine` | 3.20 | Runtime base |

### Features
- **Parallel scanning**: All three tools run concurrently for faster results
- **Platform-aware**: Default `linux/amd64`, configurable via `--platform`
- **Unified scoring**: Combined risk score from all tools (A-F grades)
- **Multiple outputs**: Table, JSON, SARIF, HTML formats
- **Graceful degradation**: Continues if one tool fails, marks result as PARTIAL
- **Local-first**: No cloud dependency, runs entirely on user's machine

### Security
- Never executes scanned images (uses `docker create` + `docker export` only)
- Safe tar extraction with path traversal protection
- Secrets redacted in reports (metadata only)
- Pinned tool versions with documented checksums

---

## Version History

### Nightly Builds (`:latest` / `:nightly`)
- Rebuilt daily at 02:00 UTC
- Automatically fetches latest versions of Trivy, Malcontent, Magika
- Best for development and testing
- Not recommended for production (use pinned versions instead)

### Release Tags (`:x.y.z`)
- Immutable, reproducible builds
- Pinned tool versions as documented above
- Recommended for production and CI/CD

---

## Upgrading

### From pre-release to 0.1.0
This is the initial release. No upgrade path needed.

### Checking Current Versions
```bash
# Check vnaiscan version
vnaiscan version

# Check bundled tool versions
vnaiscan tools
```

---

## Links

- [GitHub Releases](https://github.com/vnai-dev/vnaiscan/releases)
- [Docker Images](https://github.com/vnai-dev/vnaiscan/pkgs/container/vnaiscan)
- [Documentation](https://vnai.dev/docs/scan)

[Unreleased]: https://github.com/vnai-dev/vnaiscan/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/vnai-dev/vnaiscan/releases/tag/v0.1.0
