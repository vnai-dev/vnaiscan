# vnaiscan Homebrew Formula
# Install: brew install vnai-dev/tap/vnaiscan

class Vnaiscan < Formula
  desc "AI Agent Image Security Scanner - Scan container images for CVEs, secrets, and malware"
  homepage "https://scan.vnai.dev"
  version "0.1.0"
  license "Apache-2.0"

  on_macos do
    on_arm do
      url "https://github.com/vnai-dev/vnaiscan/releases/download/v#{version}/vnaiscan_#{version}_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    end
    on_intel do
      url "https://github.com/vnai-dev/vnaiscan/releases/download/v#{version}/vnaiscan_#{version}_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/vnai-dev/vnaiscan/releases/download/v#{version}/vnaiscan_#{version}_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    end
    on_intel do
      url "https://github.com/vnai-dev/vnaiscan/releases/download/v#{version}/vnaiscan_#{version}_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end
  end

  depends_on "trivy"
  
  # Note: malcontent and magika don't have Homebrew formulas yet
  # Users can install them manually or use Docker mode

  def install
    bin.install "vnaiscan"
  end

  def caveats
    <<~EOS
      vnaiscan requires additional tools for full functionality:

      Trivy (installed via Homebrew dependency):
        ✓ Already installed

      Malcontent (run via Docker):
        For malcontent support, Docker is required.
        Or build from source: https://github.com/chainguard-dev/malcontent

      Magika (install via Cargo):
        cargo install --locked magika-cli

      Alternatively, use Docker for a complete setup:
        docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \\
          ghcr.io/vnai-dev/vnaiscan scan <image:tag>
    EOS
  end

  test do
    assert_match "vnaiscan version", shell_output("#{bin}/vnaiscan --version")
  end
end
