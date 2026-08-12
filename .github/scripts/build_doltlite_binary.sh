#!/usr/bin/env bash
set -euo pipefail

# Builds a DoltLite-enabled dolt-mcp-server binary for the *native* platform
# and packages it into a tar.gz archive. DoltLite builds require cgo, so
# unlike build_binaries.sh this script cannot cross-compile: run it on a
# runner that matches the target platform.
#
# DoltLite is built from its pinned source tag on the native
# runner. This preserves the runner's deployment target (especially on
# macOS) instead of inheriting the minimum OS version of a prebuilt archive.
#
# Usage:
#   build_doltlite_binary.sh <doltlite_platform> <archive_platform> <output_directory>
#
#   doltlite_platform: platform component of the doltlite release asset name
#                      (e.g. linux-x64, osx-arm64)
#   archive_platform:  platform suffix for the output archive
#                      (e.g. linux-amd64, darwin-arm64)
#
# Required environment:
#   DOLTLITE_VERSION: doltlite release tag to build against (e.g. v0.11.45)
#
# Output:
#   <output_directory>/dolt-mcp-server-doltlite-<archive_platform>.tar.gz

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

if [[ $# -lt 3 || -z "${1:-}" || -z "${2:-}" || -z "${3:-}" ]]; then
  echo "Usage: $(basename "$0") <doltlite_platform> <archive_platform> <output_directory>" >&2
  exit 2
fi

DOLTLITE_PLATFORM="$1"
ARCHIVE_PLATFORM="$2"
OUT_DIR="$3"

: "${DOLTLITE_VERSION:?DOLTLITE_VERSION environment variable is required (e.g. v0.11.45)}"
DOLTLITE_VERSION_NUM="${DOLTLITE_VERSION#v}"

mkdir -p "$OUT_DIR" staging
OUT_DIR_ABS="$(cd "$OUT_DIR" && pwd)"

# Download and build the pinned DoltLite autoconf source release.
SOURCE_PARENT="$(mktemp -d)"
SOURCE_NAME="doltlite-${DOLTLITE_VERSION_NUM}"
SOURCE_URL="https://github.com/dolthub/doltlite/archive/refs/tags/${DOLTLITE_VERSION}.tar.gz"
echo "Downloading ${SOURCE_URL}"
curl -fsSL -o "${SOURCE_PARENT}/doltlite.tar.gz" "$SOURCE_URL"
tar -xzf "${SOURCE_PARENT}/doltlite.tar.gz" -C "$SOURCE_PARENT"

SOURCE_DIR="${SOURCE_PARENT}/${SOURCE_NAME}"
LIB_DIR="${SOURCE_PARENT}/build"
mkdir -p "$LIB_DIR"
(
  cd "$LIB_DIR"
  "${SOURCE_DIR}/configure"
  jobs=2
  if command -v nproc >/dev/null 2>&1; then
    jobs="$(nproc)"
  elif command -v sysctl >/dev/null 2>&1; then
    jobs="$(sysctl -n hw.ncpu)"
  fi
  make -j"$jobs" doltlite-lib
)

if [[ ! -f "${LIB_DIR}/libdoltlite.a" || ! -f "${LIB_DIR}/sqlite3.h" ]]; then
  echo "Error: DoltLite source build did not produce libdoltlite.a and sqlite3.h" >&2
  exit 1
fi

# mattn/go-sqlite3 adds -lsqlite3 when built with the libsqlite3 tag. Point
# that name at the same static DoltLite archive.
cp "${LIB_DIR}/libdoltlite.a" "${LIB_DIR}/libsqlite3.a"

EXTRA_LIBS="-lz -lpthread"
if [[ "$(uname -s)" == "Linux" ]]; then
  EXTRA_LIBS="${EXTRA_LIBS} -lm -ldl"
fi

name="dolt-mcp-server-doltlite-${ARCHIVE_PLATFORM}.tar.gz"
echo "Building doltlite/${ARCHIVE_PLATFORM} -> ${OUT_DIR_ABS}/${name}"
rm -f staging/dolt-mcp-server-doltlite
CGO_ENABLED=1 \
  CGO_CFLAGS="-I${LIB_DIR}" \
  CGO_LDFLAGS="-L${LIB_DIR} ${EXTRA_LIBS}" \
  go build -trimpath -tags "doltlite libsqlite3" -ldflags "-s -w" \
  -o staging/dolt-mcp-server-doltlite ./mcp/cmd/dolt-mcp-server
tar -C staging -czf "${OUT_DIR_ABS}/${name}" dolt-mcp-server-doltlite
rm -f staging/dolt-mcp-server-doltlite
