package settings

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"os"
	"regexp"
	"strings"
	"time"
	"io"

	"golang.org/x/exp/slog"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	settingsState "github.com/osuAkatsuki/hanayo/app/states/settings"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

var hexColourRegex = regexp.MustCompile("^#([a-fA-F0-9]{6}|[a-fA-F0-9]{3})$")

func ProfileBackgroundSubmitHandler(c *gin.Context) {
	settings := settingsState.GetSettings()
	ctx := sessions.GetContext(c)
	if ctx.User.ID == 0 {
		tu.Resp403(c)
		return
	}

	var m msg.Message = msg.SuccessMessage{lu.T(c, "Your profile background has been saved.")}
	defer func() {
		sessions.AddMessage(c, m)
		sessions.GetSession(c).Save()
		c.Redirect(302, "/settings/profbackground")
	}()
	if ok, _ := services.CSRF.Validate(ctx.User.ID, c.PostForm("csrf")); !ok {
		m = msg.ErrorMessage{lu.T(c, "Your session has expired. Please try redoing what you were trying to do.")}
		return
	}
	t := c.Param("type")
	switch t {
	case "0":
		services.DB.Exec("DELETE FROM profile_backgrounds WHERE uid = ?", ctx.User.ID)
	case "1":
		// image
		file, _, err := c.Request.FormFile("value")
		if err != nil {
			m = msg.ErrorMessage{lu.T(c, "An error occurred.")}
			return
		}
		img, _, err := image.Decode(file)
		if err != nil {
			m = msg.ErrorMessage{lu.T(c, "An error occurred.")}
			return
		}
		//img = resize.Resize(1127, 250, img, resize.Bilinear)
		f, err := os.CreateTemp("", fmt.Sprintf("%d.jpg", ctx.User.ID))
		if err != nil {
			m = msg.ErrorMessage{lu.T(c, "An error occurred.")}
			c.Error(err)
			return
		}
		defer os.Remove(f.Name())

		defer f.Close()
		if err != nil {
			m = msg.ErrorMessage{lu.T(c, "An error occurred.")}
			c.Error(err)
			slog.ErrorContext(c, err.Error())
			return
		}
		err = jpeg.Encode(f, img, &jpeg.Options{
			Quality: 88,
		})
		if err != nil {
			m = msg.ErrorMessage{lu.T(c, "We were not able to save your profile background.")}
			c.Error(err)
			slog.ErrorContext(c, err.Error())
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
			Key:         aws.String(fmt.Sprintf("profile-backgrounds/%d.jpg", ctx.User.ID)),
			Body:        f,
			ContentType: aws.String("image/jpg"),
			// TODO: CacheControl?
		})
		if err != nil {
			m = msg.ErrorMessage{lu.T(c, "We were not able to save your profile background.")}
			c.Error(err)
			slog.ErrorContext(c, err.Error())
			return
		}

		saveProfileBackground(ctx, 1, fmt.Sprintf("%d.jpg", ctx.User.ID))
	case "2":
		// solid colour
		col := strings.ToLower(c.PostForm("value"))
		// verify it's valid
		if !hexColourRegex.MatchString(col) {
			m = msg.ErrorMessage{lu.T(c, "Colour is invalid")}
			return
		}
		saveProfileBackground(ctx, 2, col)
	case "3":
		// gifs
		file, _, err := c.Request.FormFile("value")
		if err != nil {
			m = msg.ErrorMessage{lu.T(c, "An error occurred.")}
			return
		}
		gifImage, err := gif.DecodeAll(file)
		if err != nil {
			m = msg.ErrorMessage{lu.T(c, "An error occurred.")}
			return
		}
		// TODO: implement resizing for gifs
		// TODO: implement gif compression

		f, err := os.CreateTemp("", fmt.Sprintf("%d.gif", ctx.User.ID))
		if err != nil {
			m = msg.ErrorMessage{lu.T(c, "An error occurred.")}
			c.Error(err)
			slog.ErrorContext(c, err.Error())
			return
		}
		defer os.Remove(f.Name())

		err = gif.EncodeAll(f, gifImage)
		if err != nil {
			m = msg.ErrorMessage{lu.T(c, "We were not able to save your profile background.")}
			c.Error(err)
			slog.ErrorContext(c, err.Error())
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
			Key:         aws.String(fmt.Sprintf("profile-backgrounds/%d.gif", ctx.User.ID)),
			Body:        f,
			ContentType: aws.String("image/gif"),
			// TODO: CacheControl?
		})
		if err != nil {
			m = msg.ErrorMessage{lu.T(c, "We were not able to save your profile background.")}
			c.Error(err)
			slog.ErrorContext(c, err.Error())
			return
		}

		saveProfileBackground(ctx, 3, fmt.Sprintf("%d.gif", ctx.User.ID))
	}

}

func saveProfileBackground(ctx models.Context, t int, val string) {
	services.DB.Exec(`INSERT INTO profile_backgrounds(uid, time, type, value)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			time = VALUES(time),
			type = VALUES(type),
			value = VALUES(value)
	`, ctx.User.ID, time.Now().Unix(), t, val)
}
