package main

import (
	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/states/services"
)

func CheckBlacklisted(c *gin.Context) {
	ip := c.ClientIP()

	rd_blacklist := services.RD.SIsMember("akatsuki:blacklist", ip)

	if rd_blacklist.Val() {
		c.Abort()
		return
	}

	c.Next()
}
