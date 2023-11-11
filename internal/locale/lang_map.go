package locale

import (
	"fmt"
	"golang.org/x/exp/slog"
	"os"
	"strings"
)

var languageMap = make(map[string]*po, 20)

func loadLanguages() {
	files, err := os.ReadDir("./locale/locales")
	if err != nil {
		slog.Error("Error reading locale directory", "error", err.Error())
		return
	}
	for _, file := range files {
		if file.Name() == "templates.pot" || file.Name() == "." || file.Name() == ".." {
			continue
		}

		p, err := parseFile("./locale/locales/" + file.Name())
		if err != nil {
			slog.Error("Error parsing po file", "error", err.Error())
			continue
		}
		if p == nil {
			slog.Error("Error parsing po file", "error", "p is nil", "file_name", file.Name())
		}

		langName := strings.TrimPrefix(strings.TrimSuffix(file.Name(), ".po"), "templates-")
		languageMap[langName] = p
	}
}

func init() {
	loadLanguages()
}

// Get retrieves a string from a language
func Get(langs []string, str string, vars ...interface{}) string {
	for _, lang := range langs {
		l := languageMap[lang]

		if l == nil {
			continue
		}

		if el := l.Translations[str]; el != "" {
			if len(vars) > 0 {

				return fmt.Sprintf(el, vars...)
			}
			return el
		}
	}

	if len(vars) > 0 {
		return fmt.Sprintf(str, vars...)
	}
	return str
}
