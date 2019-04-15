package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
)

func avatarSubmit(c *gin.Context) {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		resp403(c)
		return
	}
	var m message
	defer func() {
		simpleReply(c, m)
	}()
	if config.AvatarsFolder == "" {
		m = errorMessage{T(c, "Changing avatar is currently not possible.")}
		return
	}
	file, _, err := c.Request.FormFile("avatar")
	if err != nil {
		m = errorMessage{T(c, "An error occurred.")}
		return
	}
	img, _, err := image.Decode(file)
	if err != nil {
		m = errorMessage{T(c, "An error occurred.")}
		return
	}
	img = resize.Thumbnail(256, 256, img, resize.Bilinear)
	f, err := os.Create(fmt.Sprintf("%s/%d.png", config.AvatarsFolder, ctx.User.ID))
	defer f.Close()
	if err != nil {
		m = errorMessage{T(c, "An error occurred.")}
		c.Error(err)
		return
	}
	err = png.Encode(f, img)
	if err != nil {
		m = errorMessage{T(c, "We were not able to save your avatar.")}
		c.Error(err)
		return
	}
	m = successMessage{T(c, "Your avatar was successfully changed. It may take some time to properly update. To force a cache refresh, you can use CTRL+F5.")}
}
