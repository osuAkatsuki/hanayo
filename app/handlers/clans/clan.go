package clans

import (
	"database/sql"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/models"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

func ClanPageHandler(c *gin.Context) {
	var (
		clanID   int
		clanName string
	)

	// ctx := getContext(c)

	i := c.Param("cid")
	if _, err := strconv.Atoi(i); err != nil {
		err := services.DB.QueryRow("SELECT id, name FROM clans WHERE name = ? LIMIT 1", i).Scan(&clanID, &clanName)
		if err != nil && err != sql.ErrNoRows {
			c.Error(err)
		}
	} else {
		err := services.DB.QueryRow("SELECT id, name FROM clans WHERE id = ? LIMIT 1", i).Scan(&clanID, &clanName)
		if err != nil && err != sql.ErrNoRows {
			c.Error(err)
		}
	}

	data := new(models.ClanData)
	data.ClanID = clanID
	defer tu.Resp(c, 200, "clansample.html", data)

	if data.ClanID == 0 {
		data.TitleBar = "Clan not found"
		data.Messages = append(data.Messages, msg.WarningMessage{lu.T(c, "That clan could not be found.")})
		return
	}

	if (sessions.GetContext(c).User.Privileges & 1) != 0 {
		if services.DB.QueryRow("SELECT 1 FROM clans WHERE id = ?", clanID).Scan(new(string)) != sql.ErrNoRows {
			var bg string
			services.DB.QueryRow("SELECT background FROM clans WHERE id = ?", clanID).Scan(&bg)
			data.BannerContent = bg
			data.BannerAbsolute = true
			data.BannerType = 1
		}
	}

	data.TitleBar = lu.T(c, "%s's Clan Page", clanName)
	data.DisableHH = true
	data.Scripts = append(data.Scripts, "/static/js/pages/clan.js")
}
