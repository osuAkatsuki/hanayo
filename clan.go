package main

import (
	"database/sql"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TODO: replace with simple ResponseInfo containing userid
type clanData struct {
	baseTemplateData
	ClanID int
}

func clanPage(c *gin.Context) {
	var (
		clanID          int
		clanName        string
	)

	// ctx := getContext(c)

	i := c.Param("cid")
	if _, err := strconv.Atoi(i); err != nil {
		err := db.QueryRow("SELECT id, name FROM clans WHERE name = ? LIMIT 1", i).Scan(&clanID, &clanName)
		if err != nil && err != sql.ErrNoRows {
			c.Error(err)
		}
	} else {
		err := db.QueryRow("SELECT id, name FROM clans WHERE id = ? LIMIT 1", i).Scan(&clanID, &clanName)
		if err != nil && err != sql.ErrNoRows {
			c.Error(err)
		}
	}
	
	data := new(clanData)
	data.ClanID = clanID
	defer resp(c, 200, "clansample.html", data)

	if data.ClanID == 0 {
		data.TitleBar = "Clan not found"
		data.Messages = append(data.Messages, warningMessage{T(c, "That clan could not be found.")})
		return
	}

	if (getContext(c).User.Privileges & 1) != 0 {
		if db.QueryRow("SELECT 1 FROM clans WHERE id = ?", clanID).Scan(new(string)) != sql.ErrNoRows {
			var bg string
			db.QueryRow("SELECT background FROM clans WHERE id = ?", clanID).Scan(&bg)
			data.KyutGrill = bg
			data.KyutGrillAbsolute = true
		}
	}

	data.TitleBar = T(c, "%s's Clan Page", clanName)
	data.DisableHH = true
	data.Scripts = append(data.Scripts, "/static/clan.js")
}