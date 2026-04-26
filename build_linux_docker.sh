#!/bin/bash
set -e

# 기본: arm64 (M-series Mac 네이티브, 빠름)
# CI 검증용: AMD64 에뮬레이션은 --amd64 플래그 사용
#   ./build_linux_docker.sh --amd64

PLATFORM="linux/arm64"
if [[ "$1" == "--amd64" ]]; then
  PLATFORM="linux/amd64"
  echo "=== AMD64 에뮬레이션 빌드 (QEMU, 느림) ==="
else
  echo "=== ARM64 네이티브 빌드 (로컬 테스트용) ==="
fi

IMAGE="jbridgego-linux-builder"
OUT_DIR="$(pwd)/build/bin"

docker build \
  --platform "$PLATFORM" \
  -f Dockerfile.linux \
  -t "$IMAGE" \
  .

mkdir -p "$OUT_DIR"
CID=$(docker create --platform "$PLATFORM" "$IMAGE")
docker cp "$CID:/app/build/bin/." "$OUT_DIR/"
docker rm "$CID"

echo ""
echo "=== 빌드 완료 → build/bin/ ==="
ls -lh "$OUT_DIR"
