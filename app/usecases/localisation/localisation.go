package localisation

import (
	"github.com/gin-gonic/gin"
	localeEngine "github.com/osuAkatsuki/hanayo/internal/locale"
)

var localeLanguages = []string{"de", "pl", "it", "es", "ru", "fr", "nl", "ro", "fi", "sv", "vi", "ko"}

// T translates a string into the language specified by the request.
func T(c *gin.Context, s string, args ...interface{}) string {
	return localeEngine.Get(GetLang(c), s, args...)
}

func GetLang(c *gin.Context) []string {
	s, _ := c.Cookie("language")
	if s != "" {
		return []string{s}
	}
	return localeEngine.ParseHeader(c.Request.Header.Get("Accept-Language"))
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
