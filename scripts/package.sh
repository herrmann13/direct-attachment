#!/usr/bin/env bash
# Packages the extension into distributable zip files and generates the
# Firefox auto-update manifest (update.json).
#
# Usage: scripts/package.sh
#
# Outputs into dist/:
#   direct-attachment-chrome-<version>.zip
#   direct-attachment-firefox-<version>.zip
#   update.json  (edit the placeholder update_link before hosting)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXT="$ROOT/extension"
DIST="$ROOT/dist"

command -v zip >/dev/null 2>&1 || { echo "error: zip is required"; exit 1; }
command -v jq  >/dev/null 2>&1 || { echo "error: jq is required"; exit 1; }

VERSION="$(jq -r .version "$EXT/manifest.json")"
GECKO_ID="$(jq -r '.browser_specific_settings.gecko.id' "$EXT/manifest.json")"

mkdir -p "$DIST"

echo "==> Packaging version $VERSION (gecko id: $GECKO_ID)"

chrome_zip="$DIST/direct-attachment-chrome-$VERSION.zip"
firefox_zip="$DIST/direct-attachment-firefox-$VERSION.zip"

rm -f "$chrome_zip" "$firefox_zip"

# Firefox MV3 uses background.scripts, so its archive is an unchanged copy of
# the source directory. Chrome MV3 requires a service worker, so only the
# Chrome archive is built from a temporary copy with an adjusted manifest.
( cd "$EXT" && zip -qr "$firefox_zip" . )

chrome_stage="$(mktemp -d)"
trap 'rm -rf "$chrome_stage"' EXIT
cp -R "$EXT"/. "$chrome_stage"/
jq '.background = {service_worker: "background.js"}' \
  "$EXT/manifest.json" > "$chrome_stage/manifest.json"
( cd "$chrome_stage" && zip -qr "$chrome_zip" . )

cat > "$DIST/update.json" <<EOF
{
  "addons": {
    "$GECKO_ID": {
      "updates": [
        {
          "version": "$VERSION",
          "update_link": "https://example.com/direct-attachment-firefox-$VERSION.xpi"
        }
      ]
    }
  }
}
EOF

echo "==> Done."
echo "    Chrome : $chrome_zip"
echo "    Firefox: $firefox_zip"
echo "    update.json written with a placeholder update_link (edit before hosting)."
