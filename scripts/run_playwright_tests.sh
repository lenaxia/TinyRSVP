#!/usr/bin/env bash
# run_playwright_tests.sh — launcher for Playwright UX tests.
#
# Sets up LD_LIBRARY_PATH for the headless Chromium's shared library deps
# (libglib, libnss, libxcb, etc.) that the playwright-go installer doesn't
# provide. On Debian/Ubuntu these come from the system; on minimal/sandboxed
# environments they must be provided separately.
#
# Usage:
#   ./scripts/run_playwright_tests.sh                # run all
#   ./scripts/run_playwright_tests.sh -run TestName   # run specific test
#
# First-time setup:
#   1. go install github.com/mxschmitt/playwright-go/cmd/playwright@latest
#   2. playwright install chromium
#   3. ./scripts/run_playwright_tests.sh    # downloads deps if needed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LIBS_DIR="${PW_LIBS_DIR:-/tmp/pwlibs/extracted}"

cd "$REPO_ROOT"

# If the shared libs aren't available, try to download them from the Debian
# mirror. This is the workaround for sandboxed environments without root.
if [ ! -d "$LIBS_DIR/usr/lib/x86_64-linux-gnu" ]; then
    echo "Playwright libs not found at $LIBS_DIR; downloading..." >&2
    bash "$SCRIPT_DIR/install_playwright_deps.sh"
fi

export LD_LIBRARY_PATH="$LIBS_DIR/usr/lib/x86_64-linux-gnu:$LIBS_DIR/lib/x86_64-linux-gnu:${LD_LIBRARY_PATH:-}"

# Run the tests with a generous timeout (browser startup is slow).
exec go test -timeout 300s -v ./tests/ux_playwright/... "$@"
