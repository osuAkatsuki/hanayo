package settings

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"

	"golang.org/x/exp/slog"

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

	f, err := os.CreateTemp("", fmt.Sprintf("%d.png", ctx.User.ID))
	if err != nil {
		m = msg.ErrorMessage{lu.T(c, "An error occurred.")}
		c.Error(err)
		return
	}
	defer os.Remove(f.Name())

	err = png.Encode(f, img)
	if err != nil {
		m = msg.ErrorMessage{lu.T(c, "We were not able to save your avatar.")}
		c.Error(err)
		return
	}
	// seek file to beginning
	f.Seek(0, io.SeekStart)

	req, err := http.NewRequest(
		"POST",
		settings.INTERNAL_AVATARS_SERVICE_BASE_URL+"/api/v1/users/"+fmt.Sprint(ctx.User.ID)+"/avatar",
		f,
	)

	if err != nil {
		m = msg.ErrorMessage{lu.T(c, "We were not able to save your avatar.")}
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		return
	}

	req.Header.Add("Authorization", "Bearer "+settings.RESTRICTED_AVATARS_SERVICE_API_KEY)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m = msg.ErrorMessage{lu.T(c, "We were not able to save your avatar.")}
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		m = msg.ErrorMessage{lu.T(c, "We were not able to save your avatar.")}
		c.Error(err)
		slog.ErrorContext(c, "Avatar service returned non-2xx status", "status_code", resp.StatusCode)
		return
	}

	m = msg.SuccessMessage{lu.T(c, "Your avatar was successfully changed. It may take some time to properly update. To force a cache refresh, you can use CTRL+F5.")}
}
