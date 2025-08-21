//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type TemplateString struct {
	Key   string
	Value string
}

func main() {
	// Extract template strings from HTML files
	templateStrings := extractTemplateStrings("web/templates")

	// Generate .pot file
	generatePotFile(templateStrings, "locale/locales/templates.pot")

	// Update existing .po files
	updatePoFiles(templateStrings)

	fmt.Println("Template extraction completed!")
}

func extractTemplateStrings(dir string) map[string]string {
	templateStrings := make(map[string]string)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		strings := extractFromFile(path)
		for k, v := range strings {
			templateStrings[k] = v
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return templateStrings
	}

	return templateStrings
}

func extractFromFile(filepath string) map[string]string {
	strings := make(map[string]string)

	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filepath, err)
		return strings
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Regex to match template strings like {{ $.T "string" }} or {{ .T "string" }}
	templateRegex := regexp.MustCompile(`\{\{\s*[\$]?\.T\s+"([^"]+)"\s*\}\}`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := templateRegex.FindAllStringSubmatch(line, -1)

		for _, match := range matches {
			if len(match) >= 2 {
				key := match[1]
				strings[key] = key
			}
		}
	}

	return strings
}

func generatePotFile(templateStrings map[string]string, outputPath string) {
	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Error creating .pot file: %v\n", err)
		return
	}
	defer file.Close()

	// Write header
	header := `# SOME DESCRIPTIVE TITLE.
# Copyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER
# This file is distributed under the same license as the PACKAGE package.
# FIRST AUTHOR <EMAIL@ADDRESS>, YEAR.
#
#, fuzzy
msgid ""
msgstr ""
"Project-Id-Version: PACKAGE VERSION\n"
"Report-Msgid-Bugs-To: \n"
"POT-Creation-Date: 2024-01-01 12:00+0000\n"
"PO-Revision-Date: YEAR-MO-DA HO:MI+ZONE\n"
"Last-Translator: FULL NAME <EMAIL@ADDRESS>\n"
"Language-Team: LANGUAGE <LL@li.org>\n"
"Language: \n"
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Plural-Forms: nplurals=INTEGER; plural=EXPRESSION;\n"

`
	file.WriteString(header)

	// Sort keys for consistent output
	var keys []string
	for k := range templateStrings {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Write template strings
	for _, key := range keys {
		value := templateStrings[key]
		entry := fmt.Sprintf("msgid \"%s\"\nmsgstr \"%s\"\n\n", escapeString(key), escapeString(value))
		file.WriteString(entry)
	}
}

func updatePoFiles(templateStrings map[string]string) {
	poDir := "locale/locales"

	entries, err := os.ReadDir(poDir)
	if err != nil {
		fmt.Printf("Error reading .po directory: %v\n", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".po") {
			continue
		}

		poPath := filepath.Join(poDir, entry.Name())
		updatePoFile(poPath, templateStrings)
	}
}

func updatePoFile(poPath string, templateStrings map[string]string) {
	// Read existing .po file
	existingStrings := readPoFile(poPath)

	// Create new .po file
	file, err := os.Create(poPath)
	if err != nil {
		fmt.Printf("Error creating .po file %s: %v\n", poPath, err)
		return
	}
	defer file.Close()

	// Write header
	header := `# SOME DESCRIPTIVE TITLE.
# Copyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER
# This file is distributed under the same license as the PACKAGE package.
# FIRST AUTHOR <EMAIL@ADDRESS>, YEAR.
#
msgid ""
msgstr ""
"Project-Id-Version: PACKAGE VERSION\n"
"Report-Msgid-Bugs-To: \n"
"POT-Creation-Date: 2024-01-01 12:00+0000\n"
"PO-Revision-Date: YEAR-MO-DA HO:MI+ZONE\n"
"Last-Translator: FULL NAME <EMAIL@ADDRESS>\n"
"Language-Team: LANGUAGE <LL@li.org>\n"
"Language: \n"
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Plural-Forms: nplurals=INTEGER; plural=EXPRESSION;\n"

`
	file.WriteString(header)

	// Sort keys for consistent output
	var keys []string
	for k := range templateStrings {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Write template strings
	for _, key := range keys {
		value := templateStrings[key]
		translation := existingStrings[key]
		if translation == "" {
			translation = value // Use original as fallback
		}

		entry := fmt.Sprintf("msgid \"%s\"\nmsgstr \"%s\"\n\n", escapeString(key), escapeString(translation))
		file.WriteString(entry)
	}
}

func readPoFile(poPath string) map[string]string {
	poStrings := make(map[string]string)

	file, err := os.Open(poPath)
	if err != nil {
		return poStrings
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var currentMsgID string
	var currentMsgStr string
	var inMsgStr bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "msgid ") {
			if currentMsgID != "" && currentMsgStr != "" {
				poStrings[currentMsgID] = currentMsgStr
			}
			currentMsgID = unescapeString(strings.TrimPrefix(line, "msgid "))
			currentMsgStr = ""
			inMsgStr = false
		} else if strings.HasPrefix(line, "msgstr ") {
			currentMsgStr = unescapeString(strings.TrimPrefix(line, "msgstr "))
			inMsgStr = true
		} else if inMsgStr && line != "" {
			currentMsgStr += unescapeString(line)
		}
	}

	// Don't forget the last entry
	if currentMsgID != "" && currentMsgStr != "" {
		poStrings[currentMsgID] = currentMsgStr
	}

	return poStrings
}

func escapeString(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

func unescapeString(s string) string {
	// Remove quotes if present
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		s = s[1 : len(s)-1]
	}
	return strings.ReplaceAll(s, "\\\"", "\"")
}
