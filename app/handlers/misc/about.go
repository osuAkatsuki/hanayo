package misc

import (
	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/models"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

var data = new(models.BaseTemplateData)

func AboutPageHandler(c *gin.Context) {
	defer tu.Resp(c, 200, "about.html", data)
	data.DisableHH = true
}
