package recovery

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"strings"
	"time"

	"golang.org/x/exp/slog"

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

const passwordResetLifetime = 30 * time.Minute

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
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}

	if common.UserPrivileges(privileges)&
		(common.UserPrivilegeNormal|common.UserPrivilegePendingVerification) == 0 {
		tu.SimpleReply(c, msg.ErrorMessage{lu.T(c, "You look pretty banned/locked here.")})
		return
	}

	// generate key
	keyBytes := make([]byte, 25)
	if _, err := rand.Read(keyBytes); err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}
	key := hex.EncodeToString(keyBytes)
	now := time.Now().UTC()

	tx, err := services.DB.Begin()
	if err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE password_recovery
		SET status = 'revoked'
		WHERE user_id = ? AND status = 'unused' AND created_at >= ?`,
		id, now.Add(-passwordResetLifetime))
	if err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}

	result, err := tx.Exec(`
		INSERT INTO password_recovery(user_id, token, created_at, status)
		VALUES (?, ?, ?, 'unused')`, id, key, now)
	if err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}
	resetID, err := result.LastInsertId()
	if err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}
	if err := tx.Commit(); err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}

	content := lu.T(c,
		"Hey <b>%s</b>!<br/><br/>Someone (<i>which we really hope was you</i>), requested a password reset for your account. In case it was you, please <a href='%s'>click here</a> to reset your password on Akatsuki.<br/>Otherwise, silently ignore this email.",
		username,
		settings.APP_BASE_URL+"/pwreset/continue?k="+key,
	)
	mailMessage := mailgun.NewMessage(
		"Akatsuki <"+settings.MAILGUN_FROM+">",
		lu.T(c, "Akatsuki password recovery instructions"),
		content,
		email,
	)
	mailMessage.SetHtml(content)
	_, _, err = services.MG.Send(mailMessage)

	if err != nil {
		if _, revokeErr := services.DB.Exec(`
			UPDATE password_recovery SET status = 'revoked'
			WHERE id = ? AND status = 'unused'`, resetID); revokeErr != nil {
			c.Error(revokeErr)
			slog.ErrorContext(c, revokeErr.Error())
		}
		c.Error(err)
		slog.ErrorContext(c, err.Error())
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

	var (
		username  string
		createdAt time.Time
	)
	switch err := services.DB.QueryRow(`
		SELECT users.username, password_recovery.created_at
		FROM password_recovery
		INNER JOIN users ON users.id = password_recovery.user_id
		WHERE password_recovery.token = ? AND password_recovery.status = 'unused'
		LIMIT 1`, k).
		Scan(&username, &createdAt); err {
	case nil:
		// move on
	case sql.ErrNoRows:
		tu.RespEmpty(c, lu.T(c, "Reset password"), msg.ErrorMessage{lu.T(c, "That key could not be found. Perhaps it expired?")})
		return
	default:
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}
	if time.Now().After(createdAt.Add(passwordResetLifetime)) {
		tu.RespEmpty(c, lu.T(c, "Reset password"), msg.ErrorMessage{lu.T(c, "That key could not be found. Perhaps it expired?")})
		return
	}

	renderResetPassword(c, username, k)
}

func PasswordResetContinueSubmitHandler(c *gin.Context) {
	// todo: check logged in
	var (
		resetID   int64
		userID    int
		username  string
		createdAt time.Time
	)
	key := c.PostForm("k")
	switch err := services.DB.QueryRow(`
		SELECT password_recovery.id, users.id, users.username, password_recovery.created_at
		FROM password_recovery
		INNER JOIN users ON users.id = password_recovery.user_id
		WHERE password_recovery.token = ? AND password_recovery.status = 'unused'
		LIMIT 1`, key).
		Scan(&resetID, &userID, &username, &createdAt); err {
	case nil:
		// move on
	case sql.ErrNoRows:
		tu.RespEmpty(c, lu.T(c, "Reset password"), msg.ErrorMessage{lu.T(c, "That key could not be found. Perhaps it expired?")})
		return
	default:
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}
	if time.Now().After(createdAt.Add(passwordResetLifetime)) {
		tu.RespEmpty(c, lu.T(c, "Reset password"), msg.ErrorMessage{lu.T(c, "That key could not be found. Perhaps it expired?")})
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
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}

	tx, err := services.DB.Begin()
	if err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		UPDATE password_recovery
		SET status = 'used', used_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'unused' AND created_at >= ?`,
		resetID, time.Now().Add(-passwordResetLifetime))
	if err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}
	if rowsAffected != 1 {
		tu.RespEmpty(c, lu.T(c, "Reset password"), msg.ErrorMessage{lu.T(c, "That key could not be found. Perhaps it expired?")})
		return
	}

	_, err = tx.Exec("UPDATE users SET password_md5 = ? WHERE id = ?", pass, userID)
	if err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}

	if err := tx.Commit(); err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
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
