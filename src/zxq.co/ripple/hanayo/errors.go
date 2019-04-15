package main

import "github.com/gin-gonic/gin"

func notFound(c *gin.Context) {
	resp(c, 404, "not_found.html", &baseTemplateData{
		TitleBar:  "Not Found",
		KyutGrill: "not_found.jpg",
	})
}

func resp500(c *gin.Context) {
	resp(c, 500, "500.html", &baseTemplateData{
		TitleBar: "Internal Server Error",
	})
}
