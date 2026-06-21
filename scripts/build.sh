#!/bin/bash
# Cross-platform compilation for proxis-c2
# Enterprise-grade build script for C2 server and agent binaries

set -euo pipefail

# Configuration
PLATFORMS=("windows/amd64" "windows/arm64" "linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64")
VERSION=${1:-"dev"}
BUILD_DIR="bin"
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Create build output directory
mkdir -p "${BUILD_DIR}"

# Build server binaries
echo "Building server binaries for version ${VERSION}..."
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"
    OUTPUT_NAME="proxis-server-${GOOS}-${GOARCH}"
    
    if [ "$GOOS" == "windows" ]; then
        OUTPUT_NAME+=".exe"
    fi
    
    echo "  Building for ${GOOS}/${GOARCH}..."
    env GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X main.version=${VERSION} -X main.buildDate=${DATE} -s -w" \
        -o "${BUILD_DIR}/${OUTPUT_NAME}" \
        ./cmd/server/
done

# Build agent binaries with build tags
echo "Building agent binaries for version ${VERSION}..."
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"
    OUTPUT_NAME="proxis-agent-${GOOS}-${GOARCH}"
    
    if [ "$GOOS" == "windows" ]; then
        OUTPUT_NAME+=".exe"
    fi
    
    echo "  Building for ${GOOS}/${GOARCH}..."
    env GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X main.version=${VERSION} -X main.buildDate=${DATE} -s -w" \
        -tags "${GOOS}" \
        -o "${BUILD_DIR}/${OUTPUT_NAME}" \
        ./cmd/agent/
done

echo "Build complete. Binaries available in ${BUILD_DIR}/"