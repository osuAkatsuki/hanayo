package settings

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/internal/bbcode"
)

func ParseBBCodeSubmitHandler(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) > 65336 {
		c.Error(err)
		c.String(200, "Error")
		return
	}
	d := bbcode.ConvertBBCodeToHTML(string(body))
	c.String(200, d)
}
