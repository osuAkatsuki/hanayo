package errors

import (
	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/models"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

func NotFoundHandler(c *gin.Context) {
	tu.Resp(c, 404, "not_found.html", &models.BaseTemplateData{
		TitleBar:      "Not Found",
		BannerContent: "not_found.jpg",
		BannerType:    1,
	})
}

func Resp500(c *gin.Context) {
	tu.Resp(c, 500, "500.html", &models.BaseTemplateData{
		TitleBar: "Internal Server Error",
	})
}
