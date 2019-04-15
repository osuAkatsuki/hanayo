package main

import (
	"github.com/gin-gonic/gin"
	"zxq.co/ripple/hanayo/modules/locale"
)

// T translates a string into the language specified by the request.
func T(c *gin.Context, s string, args ...interface{}) string {
	return locale.Get(getLang(c), s, args...)
}

func (b *baseTemplateData) T(s string, args ...interface{}) string {
	return T(b.Gin, s, args...)
}

func getLang(c *gin.Context) []string {
	s, _ := c.Cookie("language")
	if s != "" {
		return []string{s}
	}
	return locale.ParseHeader(c.Request.Header.Get("Accept-Language"))
}
