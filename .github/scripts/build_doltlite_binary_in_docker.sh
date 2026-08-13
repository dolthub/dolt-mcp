#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 3 ]]; then
  echo "Usage: $(basename "$0") <doltlite_platform> <archive_platform> <output_directory>" >&2
  exit 2
fi

: "${DOLTLITE_VERSION:?DOLTLITE_VERSION environment variable is required}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
mkdir -p "$3"
OUT_DIR="$(cd "$3" && pwd)"
IMAGE="dolt-mcp-doltlite-builder:${DOLTLITE_VERSION#v}"

docker build --pull -t "$IMAGE" -f "$SCRIPT_DIR/Dockerfile.doltlite-release" "$SCRIPT_DIR"
docker run --rm \
  -e DOLTLITE_VERSION \
  -e GITHUB_WORKSPACE=/src \
  -v "$REPO_ROOT:/src" \
  -v "$OUT_DIR:/out" \
  "$IMAGE" "$1" "$2" /out
