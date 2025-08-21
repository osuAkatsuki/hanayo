#!/bin/bash

# Standard gettext-based template extraction script
# This follows the typical workflow for .po file generation

set -e

echo "Extracting template strings using xgettext..."

# Create a temporary directory for processing
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Step 1: Convert Go templates to a format xgettext can understand
echo "Converting Go templates to xgettext-compatible format..."

# Create a mapping file to track original file paths
> "$TEMP_DIR/file_mapping.txt"

find web/templates -name "*.html" -type f | while read -r file; do
    # Create a temporary file with converted syntax, preserving directory structure
    rel_path="${file#web/templates/}"
    temp_file="$TEMP_DIR/$rel_path"
    temp_dir="$(dirname "$temp_file")"

    # Create directory structure in temp dir
    mkdir -p "$temp_dir"

    # Convert {{ $.T "string" }} to _("string") for xgettext
    # Convert {{ .T "string" }} to _("string") for xgettext
    sed 's/{{ \$\.T "\([^"]*\)" }}/_("\1")/g; s/{{ \.T "\([^"]*\)" }}/_("\1")/g' "$file" > "$temp_file"

    # Record mapping for later reference correction
    echo "$temp_file:$file" >> "$TEMP_DIR/file_mapping.txt"
done

# Step 2: Use xgettext to extract strings
echo "Extracting strings with xgettext..."

# Use find to get all converted files
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

# Step 3: Fix file references to point to original files instead of temp files
echo "Fixing file references..."

sed "s|$TEMP_DIR/||g" "$TEMP_DIR/templates.pot" > "locale/locales/templates.pot"

echo "Generated templates.pot"

# Step 4: Update existing .po files with new strings
echo "Updating .po files..."

for po_file in locale/locales/templates-*.po; do
    if [ -f "$po_file" ]; then
        echo "Updating $po_file..."
        msgmerge \
            --update \
            --backup=none \
            "$po_file" \
            "locale/locales/templates.pot"
    fi
done

echo "Template extraction completed!"
echo "Updated files:"
echo "  - locale/locales/templates.pot"
echo "  - locale/locales/templates-*.po"
