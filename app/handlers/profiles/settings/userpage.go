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
		c.String(200, "Userpage content is too long, maximum is 65535 characters")
		return
	}
	d := bbcode.ConvertBBCodeToHTML(string(body))
	c.String(200, d)
}
