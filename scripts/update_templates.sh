#!/bin/bash

echo "Extracting template strings from HTML files..."
go run scripts/extract_templates.go

echo "Template extraction completed!"
echo "Updated files:"
echo "  - locale/locales/templates.pot"
echo "  - locale/locales/templates-*.po"
