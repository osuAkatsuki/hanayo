#!/bin/bash

echo "Extracting template strings from HTML files..."
./scripts/extract_templates_standard.sh

echo "Template extraction completed!"
echo "Updated files:"
echo "  - locale/locales/templates.pot"
echo "  - locale/locales/templates-*.po"
