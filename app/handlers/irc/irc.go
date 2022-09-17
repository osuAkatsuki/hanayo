package irc

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/akatsuki-api/common"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	"github.com/osuAkatsuki/hanayo/app/usecases/auth/cryptography"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

func IrcGenTokenSubmitHandler(c *gin.Context) {
	ctx := sessions.GetContext(c)
	if ctx.User.ID == 0 {
		tu.Resp403(c)
		return
	}

	services.DB.Exec("DELETE FROM irc_tokens WHERE userid = ?", ctx.User.ID)

	var s, m string
	for {
		s = common.RandomString(32)
		m = cryptography.MakeMD5(s)
		if services.DB.QueryRow("SELECT 1 FROM irc_tokens WHERE token = ? LIMIT 1", m).
			Scan(new(int)) == sql.ErrNoRows {
			break
		}
	}

	services.DB.Exec("INSERT INTO irc_tokens(userid, token) VALUES (?, ?)", ctx.User.ID, m)
	tu.Simple(c, tu.GetSimple("/irc"), []msg.Message{msg.SuccessMessage{
		lu.T(c, "Your new IRC token is <code>%s</code>. The old IRC token is not valid anymore.<br>Keep it safe, don't show it around, and store it now! We won't show it to you again.", s),
	}}, nil)
}
