#!/usr/bin/env bash
set -euo pipefail

# Builds cross-platform dolt-mcp-server archives into ./out using a staging dir.
# Outputs:
#   - out/dolt-mcp-server-linux-amd64.tar.gz
#   - out/dolt-mcp-server-linux-arm64.tar.gz
#   - out/dolt-mcp-server-darwin-amd64.tar.gz
#   - out/dolt-mcp-server-darwin-arm64.tar.gz
#   - out/dolt-mcp-server-windows-amd64.zip

# Determine repository root reliably to support running from outside .github/
if [[ -n "${GITHUB_WORKSPACE:-}" && -d "${GITHUB_WORKSPACE}" ]]; then
  REPO_ROOT="${GITHUB_WORKSPACE}"
elif git_root=$(git rev-parse --show-toplevel 2>/dev/null); then
  REPO_ROOT="$git_root"
else
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
fi
cd "$REPO_ROOT"

mkdir -p out staging

package_tgz() {
  local goos="$1"; shift
  local goarch="$1"; shift
  local name="dolt-mcp-server-${goos}-${goarch}.tar.gz"
  echo "Building ${goos}/${goarch} -> out/${name}"
  rm -f staging/dolt-mcp-server
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
    go build -trimpath -ldflags "-s -w" -o staging/dolt-mcp-server ./mcp/cmd/dolt-mcp-server
  tar -C staging -czf "out/${name}" dolt-mcp-server
  rm -f staging/dolt-mcp-server
}

package_zip_windows() {
  local goarch="$1"; shift
  local name="dolt-mcp-server-windows-${goarch}.zip"
  echo "Building windows/${goarch} -> out/${name}"
  rm -f staging/dolt-mcp-server.exe
  GOOS=windows GOARCH="$goarch" CGO_ENABLED=0 \
    go build -trimpath -ldflags "-s -w" -o staging/dolt-mcp-server.exe ./mcp/cmd/dolt-mcp-server
  (cd staging && zip -q "../out/${name}" dolt-mcp-server.exe)
  rm -f staging/dolt-mcp-server.exe
}

package_tgz linux amd64
package_tgz linux arm64
package_tgz darwin amd64
package_tgz darwin arm64
package_zip_windows amd64
