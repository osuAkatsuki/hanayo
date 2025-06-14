package settings

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/internal/bbcode"
)

const maxUserpageSize = 65335+1 // 64 KiB

func ParseBBCodeSubmitHandler(c *gin.Context) {
	reader := io.LimitReader(c.Request.Body, maxUserpageSize)
	body, err := io.ReadAll(reader)
	if err != nil || len(body) >= maxUserpageSize {
		// c.Error(err)
		c.String(200, "Error: userpage too long")
		return
	}
	d := bbcode.ConvertBBCodeToHTML(string(body))
	c.String(200, d)
}
