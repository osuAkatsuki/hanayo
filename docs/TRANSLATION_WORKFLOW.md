# Translation Workflow

This document describes how to work with translatable strings in Hanayo.

## Overview

Hanayo uses the standard GNU gettext workflow for internationalization. Template strings are extracted from HTML files and managed through `.po` files for each supported language.

## Supported Languages

- German (de)
- Spanish (es)
- Finnish (fi)
- French (fr)
- Italian (it)
- Korean (ko)
- Dutch (nl)
- Polish (pl)
- Romanian (ro)
- Russian (ru)
- Swedish (sv)
- Vietnamese (vi)

## For Developers

### Adding Translatable Strings

When adding new text to HTML templates, wrap it in the translation function:

```html
<!-- Before -->
<span>Plain text that needs translation</span>

<!-- After -->
<span>{{ $.T "Plain text that needs translation" }}</span>
```

**Supported syntax:**

- `{{ $.T "string" }}` - Most common
- `{{ .T "string" }}` - Alternative syntax

### Extracting Template Strings

After adding new translatable strings, run the extraction script:

```bash
./scripts/update_templates.sh
```

This will:

1. Extract all `{{ $.T "string" }}` patterns from HTML templates
2. Generate `locale/locales/templates.pot` with new strings
3. Update all language `.po` files with new entries
4. Mark similar existing translations as "fuzzy" for review

### File Structure

```
locale/locales/
├── templates.pot          # Template file (source of truth)
├── templates-de.po        # German translations
├── templates-es.po        # Spanish translations
├── templates-fi.po        # Finnish translations
└── ...                    # Other languages
```

## For Translators

### Understanding .po Files

`.po` files contain translation entries in this format:

```pot
#: clansample.html:231
msgid "View on Admin Panel"
msgstr "Auf Admin-Panel anzeigen"
```

- `#:` - File and line number reference
- `msgid` - Original English string
- `msgstr` - Translated string

### Fuzzy Entries

When the source text changes slightly, `msgmerge` marks translations as "fuzzy":

```pot
#: clansample.html:231
#, fuzzy
msgid "View on Admin Panel"
msgstr "View on Admin Panel"
```

**What to do:**

1. Review the new `msgid` text
2. Update the `msgstr` with proper translation
3. Remove the `#, fuzzy` line
4. Save the file

### Translation Guidelines

1. **Keep the same meaning** - Don't change the intent
2. **Maintain formatting** - Preserve HTML tags if present
3. **Use consistent terminology** - Be consistent across the application
4. **Test your translations** - Ensure they fit in the UI
5. **Remove fuzzy flags** - Only remove `#, fuzzy` after reviewing

### Submitting Translations

1. Edit the appropriate `.po` file (e.g., `templates-de.po` for German)
2. Translate all `msgstr` entries
3. Remove all `#, fuzzy` flags
4. Submit a pull request with your changes

## Technical Details

### Extraction Process

The extraction script (`scripts/extract_templates_standard.sh`) works as follows:

1. **Preprocessing**: Converts Go template syntax to xgettext-compatible format
2. **Extraction**: Uses `xgettext` to extract translatable strings
3. **Post-processing**: Fixes file references and generates proper headers
4. **Merging**: Uses `msgmerge` to update existing `.po` files

### Scripts

- `scripts/update_templates.sh` - Main script for developers
- `scripts/extract_templates_standard.sh` - Core extraction logic

### File References

File references in `.po` files point to the original template files:

```pot
#: navbar.html:107
msgid "Login"
msgstr "Anmelden"
```

This helps translators understand the context of each string.

## Common Issues

### "Unterminated string literal" Warnings

These warnings from xgettext are harmless - they occur because xgettext doesn't understand HTML structure, but it still extracts the strings correctly.

### Fuzzy Entries After Updates

This is normal behavior. When template strings change, existing translations are marked as fuzzy to ensure they're reviewed for accuracy.

### Missing Translations

If a `msgstr` is empty, the application will fall back to the English `msgid`. This is intentional for incomplete translations.

## Best Practices

### For Developers

1. **Use descriptive strings** - Make translation context clear
2. **Extract regularly** - Run the script after adding new strings
3. **Test with translations** - Verify UI works with different languages
4. **Commit .po files** - Include updated translation files in commits

### For Translators

1. **Review fuzzy entries** - Always check why a translation was marked fuzzy
2. **Maintain consistency** - Use the same terms for similar concepts
3. **Consider context** - Look at the file reference to understand usage
4. **Test your work** - Verify translations work in the application

## Getting Help

If you encounter issues with the translation workflow:

1. Check this documentation
2. Look at existing `.po` files for examples
3. Open an issue on GitHub
4. Ask in the project's Discord server

## Contributing

We welcome translation contributions! Please:

1. Fork the repository
2. Create a branch for your translations
3. Update the appropriate `.po` file
4. Submit a pull request

Thank you for helping make Hanayo accessible to users worldwide! 🌍
