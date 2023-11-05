package settings

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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

	sess := session.Must(session.NewSession(&aws.Config{
		Region:   aws.String(settings.AWS_REGION),
		Endpoint: aws.String(settings.AWS_ENDPOINT_URL),
	}))
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(settings.AWS_BUCKET_NAME),
		Key:         aws.String(fmt.Sprintf("avatars/%d.png", ctx.User.ID)),
		Body:        f,
		ContentType: aws.String("image/png"),
		// TODO: CacheControl?
	})
	if err != nil {
		m = msg.ErrorMessage{lu.T(c, "We were not able to save your avatar.")}
		c.Error(err)
		return
	}

	m = msg.SuccessMessage{lu.T(c, "Your avatar was successfully changed. It may take some time to properly update. To force a cache refresh, you can use CTRL+F5.")}
}
