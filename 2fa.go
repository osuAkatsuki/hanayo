package main

import "github.com/gin-gonic/gin"

func clientIP(c *gin.Context) string {
	ff := c.Request.Header.Get("CF-Connecting-IP")
	if ff != "" {
		return ff
	}
	return c.ClientIP()
}