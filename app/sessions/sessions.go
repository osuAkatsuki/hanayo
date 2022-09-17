package sessions

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/akatsuki-api/common"
	"github.com/osuAkatsuki/hanayo/app/models"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	"github.com/osuAkatsuki/hanayo/app/usecases/auth/cryptography"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	su "github.com/osuAkatsuki/hanayo/app/usecases/sessions"
)

func SessionInitializer() func(c *gin.Context) {
	return func(c *gin.Context) {
		sess := sessions.Default(c)

		var ctx models.Context

		var passwordChanged bool
		userid := sess.Get("userid")
		if userid, ok := userid.(int); ok {
			ctx.User.ID = userid
			var (
				pRaw     int64
				password string
			)
			err := services.DB.QueryRow("SELECT username, privileges, flags, password_md5 FROM users WHERE id = ?", userid).
				Scan(&ctx.User.Username, &pRaw, &ctx.User.Flags, &password)

			if err != nil {
				c.Error(err)
			}

			if sess.Get("logout") == nil {
				sess.Set("logout", common.RandomString(15))
			}

			ctx.User.Privileges = common.UserPrivileges(pRaw)
			services.DB.Exec("UPDATE users SET latest_activity = ? WHERE id = ?", time.Now().Unix(), userid)
			if s, ok := sess.Get("pw").(string); !ok || cryptography.MakeMD5(password) != s {
				ctx = models.Context{}
				sess.Clear()
				passwordChanged = true
			}
		}

		if ctx.User.ID != 0 {
			tok := sess.Get("token")
			if tok, ok := tok.(string); ok {
				ctx.Token = tok
			}
			oldToken := ctx.Token
			ctx.Token, _ = su.CheckToken(ctx.Token, ctx.User.ID, c)
			// Set rt cookie in case:
			// - User has not got a token in rt
			// - Token has been updated with checkToken
			// - user still has old token in rt
			if x, _ := c.Cookie("rt"); oldToken != ctx.Token || x != ctx.Token {
				http.SetCookie(c.Writer, &http.Cookie{
					Name:    "rt",
					Value:   ctx.Token,
					Expires: time.Now().Add(time.Hour * 24 * 30 * 1),
				})
				sess.Set("token", ctx.Token)
			}
		}

		var addBannedMessage bool
		if ctx.User.ID != 0 && (ctx.User.Privileges&common.UserPrivilegeNormal == 0) {
			ctx = models.Context{}
			sess.Clear()
			addBannedMessage = true
		}

		ctx.Language = lu.GetLanguageFromGin(c)

		c.Set("context", ctx)
		c.Set("session", sess)

		if addBannedMessage {
			AddMessage(c, msg.WarningMessage{lu.T(c, "You have been automatically logged out of your account because your account has either been banned or locked. Should you believe this is a mistake, you can contact our support team on <a href='/discord'>Discord</a>.")})
		}
		if passwordChanged {
			AddMessage(c, msg.WarningMessage{lu.T(c, "You have been automatically logged out for security reasons. Please <a href='/login?redir=%s'>log back in</a>.", url.QueryEscape(c.Request.URL.Path))})
		}

		c.Next()
	}
}

func AddMessage(c *gin.Context, m msg.Message) {
	sess := GetSession(c)
	var messages []msg.Message
	messagesRaw := sess.Get("messages")
	if messagesRaw != nil {
		messages = messagesRaw.([]msg.Message)
	}
	messages = append(messages, m)
	sess.Set("messages", messages)
}

func GetMessages(c *gin.Context) []msg.Message {
	sess := GetSession(c)
	messagesRaw := sess.Get("messages")
	if messagesRaw == nil {
		return nil
	}
	sess.Delete("messages")
	return messagesRaw.([]msg.Message)
}

func GetSession(c *gin.Context) sessions.Session {
	return c.MustGet("session").(sessions.Session)
}

func GetContext(c *gin.Context) models.Context {
	return c.MustGet("context").(models.Context)
}
