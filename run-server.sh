#!/bin/bash
set -e

cd "$(dirname "$0")"

echo "Building frontend assets..."
(cd web && gulp)

echo "Building Go binary..."
go build -o hanayo .

if [[ "$(uname)" == "Darwin" ]]; then
    if ! codesign --verify ./hanayo 2>/dev/null; then
        echo "Re-signing executable for macOS..."
        codesign --force --sign - ./hanayo
    fi
fi

echo "Starting server..."
./hanayo
