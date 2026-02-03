# syntax=docker/dockerfile:1

# vnaiscan - AI Agent Image Security Scanner
# Multi-tool scanner with Trivy, Malcontent, and Magika

# =============================================================================
# Stage 1: Build vnaiscan Go binary
# =============================================================================
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /vnaiscan ./cmd/vnaiscan

# =============================================================================
# Stage 2: Build Magika from source (no pre-built ARM64 binaries available)
# =============================================================================
FROM rust:1.87-slim AS magika-builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

RUN cargo install --locked magika-cli \
    && cp /usr/local/cargo/bin/magika /magika

# =============================================================================
# Stage 3: Final image with all security tools
# Using Ubuntu for better glibc compatibility (Magika requires glibc >= 2.38)
# =============================================================================
FROM ubuntu:24.04

LABEL org.opencontainers.image.title="vnaiscan"
LABEL org.opencontainers.image.description="AI Agent Image Security Scanner"
LABEL org.opencontainers.image.source="https://github.com/vnai-dev/vnaiscan"
LABEL org.opencontainers.image.vendor="vnai.dev"

# Prevent interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install base dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    docker.io \
    tar \
    && rm -rf /var/lib/apt/lists/*

# -----------------------------------------------------------------------------
# Install Trivy
# https://github.com/aquasecurity/trivy/releases
# -----------------------------------------------------------------------------
ARG TRIVY_VERSION=0.69.0
RUN curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin v${TRIVY_VERSION}

# Pre-download Trivy DB (optional, makes first scan faster)
# RUN trivy image --download-db-only

# -----------------------------------------------------------------------------
# Install Malcontent from Chainguard container image (no prebuilt binaries available)
# https://github.com/chainguard-dev/malcontent
# https://images.chainguard.dev/directory/image/malcontent/overview
# -----------------------------------------------------------------------------
# Note: Malcontent requires YARA-X which is complex to build, so we copy from their official image
COPY --from=cgr.dev/chainguard/malcontent:latest /usr/bin/mal /usr/local/bin/mal
COPY --from=cgr.dev/chainguard/malcontent:latest /usr/lib/libyara_x_capi.so.1 /usr/lib/libyara_x_capi.so.1

# -----------------------------------------------------------------------------
# Install Magika (built from source in magika-builder stage)
# https://github.com/google/magika
# -----------------------------------------------------------------------------
COPY --from=magika-builder /magika /usr/local/bin/magika

# -----------------------------------------------------------------------------
# Copy vnaiscan binary
# -----------------------------------------------------------------------------
COPY --from=builder /vnaiscan /usr/local/bin/vnaiscan

# Create non-root user (UID 1001 to avoid conflicts with existing users)
RUN useradd -m -u 1001 scanner
USER scanner

WORKDIR /workspace

ENTRYPOINT ["vnaiscan"]
CMD ["--help"]
