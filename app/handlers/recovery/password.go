package recovery

import (
	"database/sql"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/akatsuki-api/common"
	eh "github.com/osuAkatsuki/hanayo/app/handlers/errors"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	settingsState "github.com/osuAkatsuki/hanayo/app/states/settings"
	au "github.com/osuAkatsuki/hanayo/app/usecases/auth"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	"github.com/osuAkatsuki/hanayo/app/usecases/misc"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
	"gopkg.in/mailgun/mailgun-go.v1"
)

func PasswordResetPageHandler(c *gin.Context) {
	settings := settingsState.GetSettings()
	ctx := sessions.GetContext(c)
	if ctx.User.ID != 0 {
		tu.SimpleReply(c, msg.ErrorMessage{lu.T(c, "You're already logged in!")})
		return
	}

	// recaptcha verify
	if settings.RECAPTCHA_SECRET_KEY != "" && !misc.RecaptchaCheck(c) {
		tu.SimpleReply(c, msg.ErrorMessage{lu.T(c, "Captcha is invalid.")})
		return
	}

	field := "username_safe"
	if strings.Contains(c.PostForm("username"), "@") {
		field = "email"
	}

	user_safe := common.SafeUsername(c.PostForm("username"))

	var (
		id         int
		username   string
		email      string
		privileges uint64
	)

	err := services.DB.QueryRow("SELECT id, username, email, privileges FROM users WHERE "+field+" = ?",
		user_safe).
		Scan(&id, &username, &email, &privileges)

	switch err {
	case nil:
		// ignore
	case sql.ErrNoRows:
		tu.SimpleReply(c, msg.ErrorMessage{lu.T(c, "That user could not be found.")})
		return
	default:
		c.Error(err)
		eh.Resp500(c)
		return
	}

	if common.UserPrivileges(privileges)&
		(common.UserPrivilegeNormal|common.UserPrivilegePendingVerification) == 0 {
		tu.SimpleReply(c, msg.ErrorMessage{lu.T(c, "You look pretty banned/locked here.")})
		return
	}

	// generate key
	key := common.RandomString(50)

	// TODO: WHY THE FUCK DOES THIS USE USERNAME AND NOT ID PLEASE WRITE MIGRATION
	_, err = services.DB.Exec("INSERT INTO password_recovery(k, u) VALUES (?, ?)", key, username)

	if err != nil {
		c.Error(err)
		eh.Resp500(c)
		return
	}

	content := lu.T(c,
		"Hey <b>%s</b>!<br/><br/>Someone (<i>which we really hope was you</i>), requested a password reset for your account. In case it was you, please <a href='%s'>click here</a> to reset your password on Akatsuki.<br/>Otherwise, silently ignore this email.",
		username,
		settings.APP_BASE_URL+"/pwreset/continue?k="+key,
	)
	mailMessage := mailgun.NewMessage(
		settings.MAILGUN_FROM,
		lu.T(c, "Akatsuki password recovery instructions"),
		content,
		email,
	)
	mailMessage.SetHtml(content)
	_, _, err = services.MG.Send(mailMessage)

	if err != nil {
		c.Error(err)
		eh.Resp500(c)
		return
	}

	sessions.AddMessage(c, msg.SuccessMessage{lu.T(c, "Done! You should shortly receive an email from us at the email you used to sign up on Akatsuki.")})
	sessions.GetSession(c).Save()
	c.Redirect(302, "/")
}

func PasswordResetContinuePageHandler(c *gin.Context) {
	k := c.Query("k")

	// todo: check logged in
	if k == "" {
		tu.RespEmpty(c, lu.T(c, "Password reset"), msg.ErrorMessage{lu.T(c, "Nope.")})
		return
	}

	var username string
	switch err := services.DB.QueryRow("SELECT u FROM password_recovery WHERE k = ? LIMIT 1", k).
		Scan(&username); err {
	case nil:
		// move on
	case sql.ErrNoRows:
		tu.RespEmpty(c, lu.T(c, "Reset password"), msg.ErrorMessage{lu.T(c, "That key could not be found. Perhaps it expired?")})
		return
	default:
		c.Error(err)
		eh.Resp500(c)
		return
	}

	renderResetPassword(c, username, k)
}

func PasswordResetContinueSubmitHandler(c *gin.Context) {
	// todo: check logged in
	var username string
	switch err := services.DB.QueryRow("SELECT u FROM password_recovery WHERE k = ? LIMIT 1", c.PostForm("k")).
		Scan(&username); err {
	case nil:
		// move on
	case sql.ErrNoRows:
		tu.RespEmpty(c, lu.T(c, "Reset password"), msg.ErrorMessage{lu.T(c, "That key could not be found. Perhaps it expired?")})
		return
	default:
		c.Error(err)
		eh.Resp500(c)
		return
	}

	p := c.PostForm("password")

	if s := au.ValidatePassword(p); s != "" {
		renderResetPassword(c, username, c.PostForm("k"), msg.ErrorMessage{lu.T(c, s)})
		return
	}

	pass, err := au.GeneratePassword(p)
	if err != nil {
		c.Error(err)
		eh.Resp500(c)
		return
	}

	_, err = services.DB.Exec("UPDATE users SET password_md5 = ?, salt = '', password_version = '2' WHERE username = ?",
		pass, username)
	if err != nil {
		c.Error(err)
		eh.Resp500(c)
		return
	}

	_, err = services.DB.Exec("DELETE FROM password_recovery WHERE k = ? LIMIT 1", c.PostForm("k"))
	if err != nil {
		c.Error(err)
		eh.Resp500(c)
		return
	}

	sessions.AddMessage(c, msg.SuccessMessage{lu.T(c, "Alright, we've changed your password, you should be able to login! Have fun!")})
	sessions.GetSession(c).Save()
	c.Redirect(302, "/login")
}

func renderResetPassword(c *gin.Context, username, k string, messages ...msg.Message) {
	tu.Simple(c, tu.GetSimpleByFilename("pwreset/continue.html"), messages, map[string]interface{}{
		"Username": username,
		"Key":      k,
	})
}
