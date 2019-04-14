package main

import (
	"database/sql"
	"strconv"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"time"
)

// TODO: replace with simple ResponseInfo containing userid
type clanData struct {
	baseTemplateData
	ClanID int
}


func leaveClan(c *gin.Context) {
	i := c.Param("cid")
	// login check
	if getContext(c).User.ID == 0 {
			resp403(c)
			return
		}
	if db.QueryRow("SELECT 1 FROM user_clans WHERE user = ? AND clan = ? AND perms = 8", getContext(c).User.ID, i).
		Scan(new(int)) == sql.ErrNoRows {
		// check if a nigga the clan
		if db.QueryRow("SELECT 1 FROM user_clans WHERE user = ? AND clan = ?", getContext(c).User.ID, i).
			Scan(new(int)) == sql.ErrNoRows {
			addMessage(c, errorMessage{T(c, "Unexpected Error...")})
			return
		}
		// idk how the fuck this gonna work but fuck it
		
			
		db.Exec("DELETE FROM user_clans WHERE user = ? AND clan = ?", getContext(c).User.ID, i)
		addMessage(c, successMessage{T(c, "Left clan.")})
		getSession(c).Save()
		c.Redirect(302, "/c/"+i)
	} else {
		//check if user even in clan!!!
		if db.QueryRow("SELECT 1 FROM user_clans WHERE user = ? AND clan = ?", getContext(c).User.ID, i).
			Scan(new(int)) == sql.ErrNoRows {
			addMessage(c, errorMessage{T(c, "Unexpected Error...")})
			return
		}
		// delete invites
		db.Exec("DELETE FROM clans_invites WHERE clan = ?", i)
		// delete all members out of clan :c
		db.Exec("DELETE FROM user_clans WHERE clan = ?", i)
		// delete clan :c
		db.Exec("DELETE FROM clans WHERE id = ?", i)
		
		addMessage(c, successMessage{T(c, "Disbanded Clan.")})
		getSession(c).Save()
		c.Redirect(302, "/clans?mode=0")
	}
	

}


func clanPage(c *gin.Context) {
	var (
		clanID           int
		clanName         string
		clanDescription  string
		clanIcon         string
	)

	// ctx := getContext(c)

	i := c.Param("cid")
	if _, err := strconv.Atoi(i); err != nil {
		err := db.QueryRow("SELECT id, name, description, icon FROM clans WHERE name = ? LIMIT 1", i).Scan(&clanID, &clanName, &clanDescription, &clanIcon)
		if err != nil && err != sql.ErrNoRows {
			c.Error(err)
		}
	} else {
		err := db.QueryRow(`SELECT id, name, description, icon FROM clans WHERE id = ? LIMIT 1`, i).Scan(&clanID, &clanName, &clanDescription, &clanIcon)
		switch {
		case err == nil:
		case err == sql.ErrNoRows:
			err := db.QueryRow("SELECT id, name, description, icon FROM clans WHERE name = ? LIMIT 1", i).Scan(&clanID, &clanName, &clanDescription, &clanIcon)
			if err != nil && err != sql.ErrNoRows {
				c.Error(err)
			}
		default:
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

	if getContext(c).User.Privileges&1 > 0 {
		if db.QueryRow("SELECT 1 FROM clans WHERE clan = ?", clanID).Scan(new(string)) != sql.ErrNoRows {
			var bg string
			db.QueryRow("SELECT background FROM clans WHERE id = ?", clanID).Scan(&bg)
			data.KyutGrill = bg
			data.KyutGrillAbsolute = true
		}
	}

	data.TitleBar = T(c, "%s's Clan Page", clanName)
	data.DisableHH = true
	data.Scripts = append(data.Scripts, "/static/profile.js")
}

func checkCount(rows *sql.Rows) (count int) {
 	for rows.Next() {
    	err:= rows.Scan(&count)
    	if err != nil {
			panic(err)
		}
    }   
    return count
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano()+int64(3))
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func createInvite(c *gin.Context) {
ctx := getContext(c)
	if string(c.PostForm("password")) == "" && string(c.PostForm("email")) == "" && string(c.PostForm("tag")) == "" && string(c.PostForm("bg")) == "" {
		
		
		if ctx.User.ID == 0 {
			resp403(c)
			return
		}
		// big perms check lol ok
		var perms int
		db.QueryRow("SELECT perms FROM user_clans WHERE user = ? AND perms = 8 LIMIT 1", ctx.User.ID).Scan(&perms)
		// delete old invite
		var clan int
		db.QueryRow("SELECT clan FROM user_clans WHERE user = ? AND perms = 8 LIMIT 1", ctx.User.ID).Scan(&clan)
		if clan == 0 {
			resp403(c)
			return
		}
		
		db.Exec("DELETE FROM clans_invites WHERE clan = ?", clan)
		
		var s string
		
		s = randSeq(8)

		db.Exec("INSERT INTO clans_invites(clan, invite) VALUES (?, ?)", clan, s)
	} else {
		// big perms check lol ok
		var perms int
		db.QueryRow("SELECT perms FROM user_clans WHERE user = ? AND perms = 8 LIMIT 1", ctx.User.ID).Scan(&perms)
		// delete old invite
		var clan int
		db.QueryRow("SELECT clan FROM user_clans WHERE user = ? AND perms = 8 LIMIT 1", ctx.User.ID).Scan(&clan)
		if clan == 0 {
			resp403(c)
			return
		}
		
		tag := "0"
		if c.PostForm("tag") != "" {
			tag = c.PostForm("tag")
		}
		
		if db.QueryRow("SELECT 1 FROM clans WHERE tag = ? AND id != ?", c.PostForm("tag"), clan).Scan(new(int)) != sql.ErrNoRows {
			resp403(c)
			addMessage(c, errorMessage{T(c, "A clan with that tag already exists...")})
			return
		}
		
		db.Exec("UPDATE clans SET description = ?, icon = ?, tag = ?, background = ? WHERE id = ?", c.PostForm("password"), c.PostForm("email"), tag, c.PostForm("bg"), clan)
	}
	addMessage(c, successMessage{T(c, "Success")})
	getSession(c).Save()
	c.Redirect(302, "/settings/clansettings")
}


func clanInvite(c *gin.Context) {
	i := c.Param("inv")
	
	res := resolveInvite(i)
	s := strconv.Itoa(res)
	if res != 0 {
	
		// check if a nigga logged in
		if getContext(c).User.ID == 0 {
			resp403(c)
			return
		}
	
		// restricted stuff
		if getContext(c).User.Privileges & 1 != 1 {
			resp403(c)
			return
		}
		
		// check if clan even exists?
			if db.QueryRow("SELECT 1 FROM clans WHERE id = ?", res).
			Scan(new(int)) == sql.ErrNoRows {

				addMessage(c, errorMessage{T(c, "Clan doesn't exist.")})
				getSession(c).Save()
				c.Redirect(302, "/c/"+s)
				return
			}
		// check if a nigga in a clan already
			if db.QueryRow("SELECT 1 FROM user_clans WHERE user = ?", getContext(c).User.ID).
			Scan(new(int)) != sql.ErrNoRows {
				
				addMessage(c, errorMessage{T(c, "Seems like you're already in a Clan")})
				getSession(c).Save()
				c.Redirect(302, "/c/"+s)
				return
			}

		// idk how the fuck this gonna work but fuck it
		var count int
		var limit int
		// members check
		db.QueryRow("SELECT COUNT(*) FROM user_clans WHERE clan = ? ", res).Scan(&count)
		db.QueryRow("SELECT mlimit FROM clans WHERE id = ? ", res).Scan(&limit)
		if count >= limit {
			addMessage(c, errorMessage{T(c, "Sorry, this clan is full.")})
			getSession(c).Save()
			c.Redirect(302, "/c/"+s)
			return
		}
		// join
		db.Exec("INSERT INTO `user_clans`(user, clan, perms) VALUES (?, ?, 1);", getContext(c).User.ID, res)
		addMessage(c, successMessage{T(c, "Joined clan.")})
		getSession(c).Save()
		c.Redirect(302, "/c/"+s)
	} else {
		resp403(c)
		addMessage(c, errorMessage{T(c, "nah nigga")})
	}
}

func clanKick(c *gin.Context) {
	if getContext(c).User.ID == 0 {
		resp403(c)
		return
	}

	if db.QueryRow("SELECT 1 FROM user_clans WHERE user = ? AND perms = 8", getContext(c).User.ID).
			Scan(new(int)) == sql.ErrNoRows {
				resp403(c)
				return
			}

	member, err := strconv.ParseInt(c.PostForm("member"), 10, 32)
	if err != nil {
		fmt.Println(err)
	}
	if member == 0 {
		resp403(c)
		return
	}

	if db.QueryRow("SELECT 1 FROM user_clans WHERE user = ? AND perms = 1", member).
			Scan(new(int)) == sql.ErrNoRows {
				resp403(c)
				return
			}

	db.Exec("DELETE FROM user_clans WHERE user = ?", member)
	addMessage(c, successMessage{T(c, "Success.")})
	getSession(c).Save()
	c.Redirect(302, "/settings/clansettings")
}

func resolveInvite(c string)(int) {
	var clanid int
	row := db.QueryRow("SELECT clan FROM clans_invites WHERE invite = ?", c)
		err := row.Scan(&clanid)
		
		if err != nil {
			fmt.Println(err)
		}
	fmt.Println(clanid)
	return clanid
}
