#!/usr/bin/env bash
set -euo pipefail

# Builds cross-platform dolt-mcp-server archives into ./out using a staging dir.
# Outputs:
#   - out/dolt-mcp-server-linux-amd64.tar.gz
#   - out/dolt-mcp-server-linux-arm64.tar.gz
#   - out/dolt-mcp-server-darwin-amd64.tar.gz
#   - out/dolt-mcp-server-darwin-arm64.tar.gz
#   - out/dolt-mcp-server-windows-amd64.zip

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Resolve repo root: prefer GITHUB_WORKSPACE, then git root, then relative to script
if [[ -n "${GITHUB_WORKSPACE:-}" && -d "${GITHUB_WORKSPACE}" ]]; then
  REPO_ROOT="${GITHUB_WORKSPACE}"
elif git_root=$(git rev-parse --show-toplevel 2>/dev/null); then
  REPO_ROOT="$git_root"
else
  REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
fi
cd "$REPO_ROOT"

# Require first argument: output directory for archives
if [[ $# -lt 1 || -z "${1:-}" ]]; then
  echo "Usage: $(basename "$0") <output_directory>" >&2
  exit 2
fi

OUT_DIR="$1"
mkdir -p "$OUT_DIR" staging
OUT_DIR_ABS="$(cd "$OUT_DIR" && pwd)"

package_tgz() {
  local goos="$1"; shift
  local goarch="$1"; shift
  local name="dolt-mcp-server-${goos}-${goarch}.tar.gz"
  echo "Building ${goos}/${goarch} -> ${OUT_DIR_ABS}/${name}"
  rm -f staging/dolt-mcp-server
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
    go build -trimpath -ldflags "-s -w" -o staging/dolt-mcp-server ./mcp/cmd/dolt-mcp-server
  tar -C staging -czf "${OUT_DIR_ABS}/${name}" dolt-mcp-server
  rm -f staging/dolt-mcp-server
}

package_zip_windows() {
  local goarch="$1"; shift
  local name="dolt-mcp-server-windows-${goarch}.zip"
  echo "Building windows/${goarch} -> ${OUT_DIR_ABS}/${name}"
  rm -f staging/dolt-mcp-server.exe
  GOOS=windows GOARCH="$goarch" CGO_ENABLED=0 \
    go build -trimpath -ldflags "-s -w" -o staging/dolt-mcp-server.exe ./mcp/cmd/dolt-mcp-server
  (cd staging && zip -q "${OUT_DIR_ABS}/${name}" dolt-mcp-server.exe)
  rm -f staging/dolt-mcp-server.exe
}

package_tgz linux amd64
package_tgz linux arm64
package_tgz darwin amd64
package_tgz darwin arm64
package_zip_windows amd64
