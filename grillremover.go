package main

import (
	"github.com/gin-gonic/gin"
)

type templateData struct {
	baseTemplateData
}

var data = new(templateData)

func aboutPage(c *gin.Context) {
	defer resp(c, 200, "about.html", data)
	data.DisableHH = true
}

func homepagePage(c *gin.Context) {
	defer resp(c, 200, "homepage.html", data)
	data.DisableHH = true
	getMessages(c)
}