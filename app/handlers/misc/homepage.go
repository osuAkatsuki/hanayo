package misc

import (
	"strconv"

	"github.com/amplitude/analytics-go/amplitude"
	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/models"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

func HomepagePageHandler(c *gin.Context) {

	data := new(models.BaseTemplateData)

	defer tu.Resp(c, 200, "homepage.html", data)

	eventOptions := amplitude.EventOptions{
		IP:       c.ClientIP(),
		Country:  c.Request.Header.Get("CF-IPCountry"),
		City:     c.Request.Header.Get("CF-IPCity"),
		Region:   c.Request.Header.Get("CF-Region"),
		Language: c.Request.Header.Get("Accept-Language"),
	}

	// include user id if logged in
	ctx := sessions.GetContext(c)
	if ctx.User.ID != 0 {
		eventOptions.UserID = strconv.Itoa(ctx.User.ID)
	}

	services.Amplitude.Track(amplitude.Event{
		EventType:       "homepage_load",
		EventOptions:    eventOptions,
		EventProperties: map[string]interface{}{"source": "hanayo"},
	})

	data.DisableHH = true
}
