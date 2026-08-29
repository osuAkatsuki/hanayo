package localisation

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	localeEngine "github.com/osuAkatsuki/hanayo/internal/locale"
)

var localeLanguages = []string{"de", "pl", "it", "es", "ru", "fr", "nl", "ro", "fi", "sv", "vi", "ko"}

// T translates a string into the language specified by the request.
func T(c *gin.Context, s string, args ...interface{}) string {
	return localeEngine.Get(GetLang(c), s, args...)
}

func getQuality(s string) float32 {
	idx := strings.Index(s, ";q=")
	if idx == -1 {
		return 1
	}

	f, err := strconv.ParseFloat(s[idx+3:], 32)
	if err != nil {
		return 1
	}
	return float32(f)
}

func parseAcceptLanguageHeader(header string) []string {
	// Parses an Accept-Language header, and sorts the values.
	if header == "" {
		return nil
	}
	parts := strings.Split(header, ",")

	sort.Slice(parts, func(i, j int) bool {
		return getQuality(parts[i]) > getQuality(parts[j])
	})

	for idx, val := range parts {
		parts[idx] = strings.Replace(strings.SplitN(val, ";q=", 2)[0], "-", "_", 1)
	}

	return parts
}

func GetLang(c *gin.Context) []string {
	s, _ := c.Cookie("language")
	if s != "" {
		return []string{s}
	}
	return parseAcceptLanguageHeader(c.Request.Header.Get("Accept-Language"))
}

func GetLanguageFromGin(c *gin.Context) string {
	for _, l := range GetLang(c) {
		for _, loc := range localeLanguages {
			if l == loc {
				return l
			}
		}
	}
	return ""
}
