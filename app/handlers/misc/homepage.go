package misc

import (
	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/models"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

func HomepagePageHandler(c *gin.Context) {

	data := new(models.BaseTemplateData)

	defer tu.Resp(c, 200, "homepage.html", data)

	data.DisableHH = true
}
