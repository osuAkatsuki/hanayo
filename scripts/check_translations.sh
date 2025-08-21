#!/bin/bash
# Translation validation script for CI
# This script validates .po files and ensures POT file is up to date
# Usage: ./scripts/check_translations.sh

set -e

echo "Checking translation files..."

# Validate all .po files
echo "1. Validating .po files..."
for po_file in locale/locales/templates-*.po; do
    if [ -f "$po_file" ]; then
        echo "Checking $po_file..."
        msgfmt --check --statistics "$po_file" -o /dev/null
    fi
done

# Check if POT file would be different from current extraction
echo "2. Checking if POT file is up to date..."
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Run extraction to temp directory
echo "Extracting template strings using xgettext..."

# Convert Go templates to xgettext-compatible format
find web/templates -name "*.html" -type f | while read -r file; do
    rel_path="${file#web/templates/}"
    temp_file="$TEMP_DIR/$rel_path"
    temp_dir="$(dirname "$temp_file")"
    mkdir -p "$temp_dir"
    sed 's/{{ \$\.T "\([^"]*\)" }}/_("\1")/g; s/{{ \.T "\([^"]*\)" }}/_("\1")/g' "$file" > "$temp_file"
done

# Extract strings
find "$TEMP_DIR" -name "*.html" -type f > "$TEMP_DIR/files_list.txt"
xgettext \
    --from-code=UTF-8 \
    --language=C \
    --keyword=_ \
    --add-comments=TRANSLATORS: \
    --package-name="Hanayo" \
    --package-version="1.0" \
    --output="$TEMP_DIR/templates.pot" \
    --files-from="$TEMP_DIR/files_list.txt"

# Fix file references
sed "s|$TEMP_DIR/||g" "$TEMP_DIR/templates.pot" > "$TEMP_DIR/extracted.pot"

# Sort for comparison
msgcat --sort-output "$TEMP_DIR/extracted.pot" -o "$TEMP_DIR/extracted_sorted.pot"
msgcat --sort-output "locale/locales/templates.pot" -o "$TEMP_DIR/current_sorted.pot"

# Compare files
if ! diff -q "$TEMP_DIR/extracted_sorted.pot" "$TEMP_DIR/current_sorted.pot" > /dev/null; then
    echo "ERROR: POT file is out of sync! Please run './scripts/update_templates.sh' and commit the changes."
    exit 1
fi

echo "✅ All translation files are valid and up to date!"
