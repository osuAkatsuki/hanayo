package main

import (
	"github.com/gin-gonic/gin"
)

func checkBlacklisted(c *gin.Context) {
	ip := c.ClientIP()

	rd_blacklist := rd.SIsMember("akatsuki:blacklist", ip)

	if rd_blacklist.Val() {
		c.Abort()
		return
	}

	c.Next()
}
