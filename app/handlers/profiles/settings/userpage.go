package settings

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/internal/bbcode"
)

func ParseBBCodeSubmitHandler(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.Error(err)
		c.String(200, "Error")
		return
	}
	d := bbcode.Compile(string(body))
	c.String(200, d)
}
