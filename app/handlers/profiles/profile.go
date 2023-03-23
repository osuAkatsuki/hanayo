package profiles

import (
	"database/sql"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/akatsuki-api/common"
	"github.com/osuAkatsuki/hanayo/app/models"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

func UserProfilePageHandler(c *gin.Context) {
	var (
		userID     int
		username   string
		privileges uint64
	)

	ctx := sessions.GetContext(c)

	u := c.Param("user")
	if _, err := strconv.Atoi(u); err != nil {
		err := services.DB.QueryRow("SELECT id, username, privileges FROM users WHERE username = ? AND "+ctx.OnlyUserPublic()+" LIMIT 1", u).Scan(&userID, &username, &privileges)
		if err != nil && err != sql.ErrNoRows {
			c.Error(err)
		}
	} else {
		err := services.DB.QueryRow(`SELECT id, username, privileges FROM users WHERE id = ? AND `+ctx.OnlyUserPublic()+` LIMIT 1`, u).Scan(&userID, &username, &privileges)
		switch {
		case err == nil:
		case err == sql.ErrNoRows:
			err := services.DB.QueryRow(`SELECT id, username, privileges FROM users WHERE username = ? AND `+ctx.OnlyUserPublic()+` LIMIT 1`, u).Scan(&userID, &username, &privileges)
			if err != nil && err != sql.ErrNoRows {
				c.Error(err)
			}
		default:
			c.Error(err)
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
			data.BannerContent = "/static/images/profbackgrounds/" + profileBackground.Value
			data.BannerAbsolute = true
			data.BannerType = profileBackground.Type
		case 2:
			data.BannerContent = profileBackground.Value
			data.BannerType = 2
		}
	}

	data.TitleBar = lu.T(c, "%s's profile", username)
	data.DisableHH = true
	data.Scripts = append(data.Scripts, "/static/js/pages/profile.js")
}
