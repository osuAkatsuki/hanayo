package main

import (
	"database/sql"
	"fmt"
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

func checkCount(rows *sql.Rows) (count int) {
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			panic(err)
		}
	}
	return count
}

func clanInvite(c *gin.Context) {
	i := c.Param("inv")

	res := resolveInvite(i)

	// Ensure invite is valid, user is logged in, and is not a restricted user.
	if res == 0 || getContext(c).User.ID == 0 || getContext(c).User.Privileges&1 != 1 {
		resp403(c)
		return
	}

	s := strconv.Itoa(res)

	// Check if clan exists.
	if db.QueryRow("SELECT 1 FROM clans WHERE id = ?", res).
		Scan(new(int)) == sql.ErrNoRows {

		addMessage(c, errorMessage{T(c, "Clan doesn't exist.")})
		getSession(c).Save()
		c.Redirect(302, "/c/"+s)
		return
	}

	// Check if user is already a member of a clan.
	if db.QueryRow("SELECT 1 FROM users WHERE id = ? AND clan_id != 0", getContext(c).User.ID).
		Scan(new(int)) != sql.ErrNoRows {

		addMessage(c, errorMessage{T(c, "Seems like you're already in a Clan")})
		getSession(c).Save()
		c.Redirect(302, "/c/"+s)
		return
	}

	// Check if clan is full.
	var count, limit int
	db.QueryRow("SELECT COUNT(id) FROM users WHERE clan_id = ?", res).Scan(&count)
	db.QueryRow("SELECT mlimit FROM clans WHERE id = ? ", res).Scan(&limit)
	if count >= limit {
		addMessage(c, errorMessage{T(c, "Sorry, this clan is full.")})
		getSession(c).Save()
		c.Redirect(302, "/c/"+s)
		return
	}

	// All checks passed.
	// Insert the user into the clan.
	db.Exec("UPDATE users SET clan_id = ?, clan_privileges = 1 WHERE id = ?", res, getContext(c).User.ID)
	//db.Exec("INSERT INTO user_clans(user, clan, perms) VALUES (?, ?, 1);", getContext(c).User.ID, res)

	addMessage(c, successMessage{T(c, "Joined clan.")})
	getSession(c).Save()
	c.Redirect(302, "/c/"+s)
}

func resolveInvite(c string) int {
	var clanid int
	//row := db.QueryRow("SELECT clan FROM clans_invites WHERE invite = ?", c)
	row := db.QueryRow("SELECT id FROM clans where invite = ?", c)
	err := row.Scan(&clanid)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(clanid)
	return clanid
}
