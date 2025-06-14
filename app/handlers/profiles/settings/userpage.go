package settings

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/internal/bbcode"
)

func ParseBBCodeSubmitHandler(c *gin.Context) {
	body := make([]byte, 65536)
	n, err := c.Request.Body.Read(body)
	if (err != nil && err != io.EOF) || n == 65536 {
		c.Error(err)
		c.String(200, "Error")
		return
	}
	d := bbcode.ConvertBBCodeToHTML(string(body))
	c.String(200, d)
}
