#!/usr/bin/env bash
# Builds the universal macOS app and stages it in dist/ under the names the
# release (.goreleaser.yaml) and the Homebrew cask download.
set -euo pipefail

cd "$(dirname "$0")/.."

name="betterglobekey-companion"
version="$(node -p "require('./package.json').version")"
bundle="src-tauri/target/universal-apple-darwin/release/bundle"

npx tauri build --target universal-apple-darwin

rm -rf dist
mkdir dist

# ditto keeps the bundle's symlinks, extended attributes, and code signature
# intact, which zip does not.
ditto -c -k --keepParent "$bundle/macos/$name.app" "dist/$name-$version-universal.zip"
cp "$bundle"/dmg/*.dmg "dist/$name-$version-universal.dmg"
