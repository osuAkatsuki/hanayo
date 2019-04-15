package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"zxq.co/ripple/rippleapi/common"
)

func ircGenToken(c *gin.Context) {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		resp403(c)
		return
	}

	db.Exec("DELETE FROM irc_tokens WHERE userid = ?", ctx.User.ID)

	var s, m string
	for {
		s = common.RandomString(32)
		m = cmd5(s)
		if db.QueryRow("SELECT 1 FROM irc_tokens WHERE token = ? LIMIT 1", m).
			Scan(new(int)) == sql.ErrNoRows {
			break
		}
	}

	db.Exec("INSERT INTO irc_tokens(userid, token) VALUES (?, ?)", ctx.User.ID, m)
	simple(c, getSimple("/irc"), []message{successMessage{
		T(c, "Your new IRC token is <code>%s</code>. The old IRC token is not valid anymore.<br>Keep it safe, don't show it around, and store it now! We won't show it to you again.", s),
	}}, nil)
}
