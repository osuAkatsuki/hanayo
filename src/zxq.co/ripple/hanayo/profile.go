package main

import (
	"database/sql"
	"strconv"

	"github.com/gin-gonic/gin"
	"zxq.co/ripple/rippleapi/common"
)

// TODO: replace with simple ResponseInfo containing userid
type profileData struct {
	baseTemplateData
	UserID int
}

func userProfile(c *gin.Context) {
	var (
		userID     int
		username   string
		privileges uint64
	)

	ctx := getContext(c)

	u := c.Param("user")
	if _, err := strconv.Atoi(u); err != nil {
		err := db.QueryRow("SELECT id, username, privileges FROM users WHERE username = ? AND "+ctx.OnlyUserPublic()+" LIMIT 1", u).Scan(&userID, &username, &privileges)
		if err != nil && err != sql.ErrNoRows {
			c.Error(err)
		}
	} else {
		err := db.QueryRow(`SELECT id, username, privileges FROM users WHERE id = ? AND `+ctx.OnlyUserPublic()+` LIMIT 1`, u).Scan(&userID, &username, &privileges)
		switch {
		case err == nil:
		case err == sql.ErrNoRows:
			err := db.QueryRow(`SELECT id, username, privileges FROM users WHERE username = ? AND `+ctx.OnlyUserPublic()+` LIMIT 1`, u).Scan(&userID, &username, &privileges)
			if err != nil && err != sql.ErrNoRows {
				c.Error(err)
			}
		default:
			c.Error(err)
		}
	}

	data := new(profileData)
	data.UserID = userID

	defer resp(c, 200, "profile.html", data)

	if data.UserID == 0 {
		data.TitleBar = "User not found"
		data.Messages = append(data.Messages, warningMessage{T(c, "That user could not be found.")})
		return
	}

	if common.UserPrivileges(privileges)&common.UserPrivilegeDonor > 0 {
		var profileBackground struct {
			Type  int
			Value string
		}
		db.Get(&profileBackground, "SELECT type, value FROM profile_backgrounds WHERE uid = ?", data.UserID)
		switch profileBackground.Type {
		case 1:
			data.KyutGrill = "/static/profbackgrounds/" + profileBackground.Value
			data.KyutGrillAbsolute = true
		case 2:
			data.SolidColour = profileBackground.Value
		}
	}

	data.TitleBar = T(c, "%s's profile", username)
	data.DisableHH = true
	data.Scripts = append(data.Scripts, "/static/profile.js")
}
