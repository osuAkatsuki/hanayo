package logout

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

func LogoutSubmitHandler(c *gin.Context) {
	ctx := sessions.GetContext(c)
	if ctx.User.ID == 0 {
		tu.RespEmpty(c, "Log out", msg.WarningMessage{lu.T(c, "You're already logged out!")})
		return
	}
	sess := sessions.GetSession(c)
	s, _ := sess.Get("logout").(string)
	if s != c.Query("k") {
		// todo: return "are you sure you want to log out?" page
		tu.RespEmpty(c, "Log out", msg.WarningMessage{lu.T(c, "Your session has expired. Please try redoing what you were trying to do.")})
		return
	}
	sess.Clear()
	http.SetCookie(c.Writer, &http.Cookie{
		Name:    "rt",
		Value:   "",
		Expires: time.Now().Add(-time.Hour),
	})
	sessions.AddMessage(c, msg.SuccessMessage{lu.T(c, "Successfully logged out.")})
	sess.Save()
	c.Redirect(302, "/")
}
