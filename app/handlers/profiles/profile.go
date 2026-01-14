package profiles

import (
	"database/sql"
	"fmt"
	"strconv"

	"golang.org/x/exp/slog"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/akatsuki-api/common"
	"github.com/osuAkatsuki/hanayo/app/models"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	settingsState "github.com/osuAkatsuki/hanayo/app/states/settings"
	"github.com/osuAkatsuki/hanayo/app/usecases/geoloc"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

func UserProfilePageHandler(c *gin.Context) {
	var (
		userID     int
		username   string
		privileges uint64
		country    string
	)

	settings := settingsState.GetSettings()
	ctx := sessions.GetContext(c)

	u := c.Param("user")
	if _, err := strconv.Atoi(u); err != nil {
		err := services.DB.QueryRow("SELECT id, username, privileges, country FROM users WHERE username = ? AND "+ctx.OnlyUserPublic()+" LIMIT 1", u).Scan(&userID, &username, &privileges, &country)
		if err != nil && err != sql.ErrNoRows {
			c.Error(err)
			slog.ErrorContext(c, err.Error())
		}
	} else {
		err := services.DB.QueryRow(`SELECT id, username, privileges, country FROM users WHERE id = ? AND `+ctx.OnlyUserPublic()+` LIMIT 1`, u).Scan(&userID, &username, &privileges, &country)
		switch {
		case err == nil:
		case err == sql.ErrNoRows:
			err := services.DB.QueryRow(`SELECT id, username, privileges, country FROM users WHERE username = ? AND `+ctx.OnlyUserPublic()+` LIMIT 1`, u).Scan(&userID, &username, &privileges, &country)
			if err != nil && err != sql.ErrNoRows {
				c.Error(err)
				slog.ErrorContext(c, err.Error())
			}
		default:
			c.Error(err)
			slog.ErrorContext(c, err.Error())
		}
	}

	data := new(models.ProfileData)
	data.UserID = userID

	defer tu.Resp(c, 200, "profile.html", data)

	if data.UserID == 0 {
		data.TitleBar = "User not found"
		data.Messages = append(data.Messages, msg.WarningMessage{lu.T(c, "That user could not be found.")})
		return
	}

	if common.UserPrivileges(privileges)&common.UserPrivilegeDonor > 0 {
		var profileBackground struct {
			Type  int
			Value string
		}
		services.DB.Get(&profileBackground, "SELECT type, value FROM profile_backgrounds WHERE uid = ?", data.UserID)
		switch profileBackground.Type {
		case 1, 3:
			data.BannerContent = fmt.Sprintf("%s/profile-backgrounds/%s", settings.PUBLIC_AVATARS_SERVICE_BASE_URL, profileBackground.Value)
			data.BannerAbsolute = true
			data.BannerType = profileBackground.Type
		case 2:
			data.BannerContent = profileBackground.Value
			data.BannerType = 2
		}
	}

	data.TitleBar = lu.T(c, "%s's profile", username)
	data.DisableHH = true
	data.Scripts = append(data.Scripts, "/static/js/pages/profile.min.js")

	// OpenGraph meta tags for social sharing
	data.OGTitle = fmt.Sprintf("%s's profile | Akatsuki", username)
	data.OGDescription = fmt.Sprintf("%s is a player from %s.", username, geoloc.CountryReadable(country))
	data.OGImage = fmt.Sprintf("%s/%d", settings.PUBLIC_AVATARS_SERVICE_BASE_URL, userID)
	data.OGUrl = fmt.Sprintf("%s/u/%d", settings.APP_BASE_URL, userID)
}
