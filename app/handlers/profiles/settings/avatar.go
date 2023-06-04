package settings

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	settingsState "github.com/osuAkatsuki/hanayo/app/states/settings"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

func AvatarSubmitHandler(c *gin.Context) {
	settings := settingsState.GetSettings()
	ctx := sessions.GetContext(c)
	if ctx.User.ID == 0 {
		tu.Resp403(c)
		return
	}
	var m msg.Message
	defer func() {
		tu.SimpleReply(c, m)
	}()
	if settings.APP_AVATAR_PATH == "" {
		m = msg.ErrorMessage{lu.T(c, "Changing avatar is currently not possible.")}
		return
	}
	file, _, err := c.Request.FormFile("avatar")
	if err != nil {
		m = msg.ErrorMessage{lu.T(c, "An error occurred.")}
		return
	}
	img, _, err := image.Decode(file)
	if err != nil {
		m = msg.ErrorMessage{lu.T(c, "An error occurred.")}
		return
	}
	img = resize.Thumbnail(256, 256, img, resize.Bilinear)

	f, err := os.Create(fmt.Sprintf("%s/%d.png", settings.APP_AVATAR_PATH, ctx.User.ID))
	defer f.Close()
	if err != nil {
		m = msg.ErrorMessage{lu.T(c, "An error occurred.")}
		c.Error(err)
		return
	}

	err = png.Encode(f, img)
	if err != nil {
		m = msg.ErrorMessage{lu.T(c, "We were not able to save your avatar.")}
		c.Error(err)
		return
	}

	m = msg.SuccessMessage{lu.T(c, "Your avatar was successfully changed. It may take some time to properly update. To force a cache refresh, you can use CTRL+F5.")}
}
