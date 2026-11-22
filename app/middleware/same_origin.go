package middleware

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func RequireSameOrigin(baseURL string) gin.HandlerFunc {
	base, err := url.Parse(baseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		panic("APP_BASE_URL must specify the website origin")
	}
	origin := base.Scheme + "://" + base.Host
	return func(c *gin.Context) {
		requestOrigin := c.GetHeader("Origin")
		if requestOrigin == "" {
			if ref, err := url.Parse(c.Request.Referer()); err == nil && ref.Host != "" {
				requestOrigin = ref.Scheme + "://" + ref.Host
			}
		}
		if requestOrigin != origin {
			c.AbortWithStatus(http.StatusForbidden)
		}
	}
}
