#!/usr/bin/env bash
# install_playwright_deps.sh — download Chromium's shared library dependencies
# from the Debian mirror and extract them into /tmp/pwlibs/extracted.
#
# Used in environments without root access where `apt-get install` and
# `playwright install-deps` both fail.
#
# After running, set LD_LIBRARY_PATH to include both:
#   - /tmp/pwlibs/extracted/usr/lib/x86_64-linux-gnu
#   - /tmp/pwlibs/extracted/lib/x86_64-linux-gnu
#
# Or use scripts/run_playwright_tests.sh which sets this up automatically.

set -euo pipefail

DEST="${PW_LIBS_DIR:-/tmp/pwlibs}"
mkdir -p "$DEST/debs" "$DEST/extracted"

cd "$DEST/debs"

# Debian bookworm (12) versions known to work with Playwright's Chromium 1.4x.
PACKAGES=(
    "g/glib2.0/libglib2.0-0_2.74.6-2+deb12u9_amd64.deb"
    "n/nspr/libnspr4_4.35-1_amd64.deb"
    "n/nss/libnss3_3.110-1+deb13u3_amd64.deb"
    "a/atk1.0/libatk1.0-0_2.36.0-2_amd64.deb"
    "a/at-spi2-atk/libatk-bridge2.0-0_2.38.0-1_amd64.deb"
    "d/dbus/libdbus-1-3_1.14.10-1~deb12u1_amd64.deb"
    "libx/libx11/libx11-6_1.8.4-2+deb12u2_amd64.deb"
    "libx/libxcomposite/libxcomposite1_0.4.5-1_amd64.deb"
    "libx/libxdamage/libxdamage1_1.1.6-1_amd64.deb"
    "libx/libxext/libxext6_1.3.4-1+b1_amd64.deb"
    "libx/libxfixes/libxfixes3_6.0.0-2_amd64.deb"
    "libx/libxrandr/libxrandr2_1.5.2-2+b1_amd64.deb"
    "libx/libxcb/libxcb1_1.15-1_amd64.deb"
    "libx/libxkbcommon/libxkbcommon0_1.5.0-1_amd64.deb"
    "a/alsa-lib/libasound2_1.2.8-1+b1_amd64.deb"
    "a/at-spi2-core/libatspi2.0-0_2.46.0-5_amd64.deb"
    "c/cups/libcups2_2.4.2-3+deb12u9_amd64.deb"
    "libd/libdrm/libdrm2_2.4.114-1+b1_amd64.deb"
    "p/pango1.0/libpango-1.0-0_1.50.12+ds-1_amd64.deb"
    "m/mesa/libgbm1_22.3.6-1+deb12u1_amd64.deb"
    "libx/libxrender/libxrender1_1.0.9-1_amd64.deb"
    "libx/libxau/libxau6_1.0.9-1_amd64.deb"
    "libx/libxdmcp/libxdmcp6_1.1.2-3_amd64.deb"
    "libx/libxi/libxi6_1.8-1+b1_amd64.deb"
    "w/wayland/libwayland-server0_1.21.0-1_amd64.deb"
)

MIRROR="http://deb.debian.org/debian/pool/main"

for pkg in "${PACKAGES[@]}"; do
    filename=$(basename "$pkg")
    if [ -f "$filename" ]; then
        continue
    fi
    if curl -sfL --max-time 30 -o "$filename" "$MIRROR/$pkg"; then
        echo "OK: $filename"
    else
        echo "FAIL: $pkg" >&2
    fi
done

for deb in *.deb; do
    dpkg-deb -x "$deb" "$DEST/extracted/"
done

echo ""
echo "Extracted to: $DEST/extracted"
echo "Set LD_LIBRARY_PATH to:"
echo "  $DEST/extracted/usr/lib/x86_64-linux-gnu:$DEST/extracted/lib/x86_64-linux-gnu"
