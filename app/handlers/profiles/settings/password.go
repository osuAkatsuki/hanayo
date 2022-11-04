package settings

import (
	"database/sql"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/akatsuki-api/common"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	au "github.com/osuAkatsuki/hanayo/app/usecases/auth"
	"github.com/osuAkatsuki/hanayo/app/usecases/auth/cryptography"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
	uu "github.com/osuAkatsuki/hanayo/app/usecases/user"
)

func ChangePasswordPageHandler(c *gin.Context) {
	ctx := sessions.GetContext(c)
	if ctx.User.ID == 0 {
		tu.Resp403(c)
	}
	s, err := services.QB.QueryRow("SELECT email FROM users WHERE id = ?", ctx.User.ID)
	if err != nil {
		c.Error(err)
	}
	tu.Simple(c, tu.GetSimpleByFilename("settings/password.html"), nil, map[string]interface{}{
		"email": s["email"],
	})
}

func ChangePasswordSubmitHandler(c *gin.Context) {
	var messages []msg.Message
	ctx := sessions.GetContext(c)
	if ctx.User.ID == 0 {
		tu.Resp403(c)
	}
	defer func() {
		s, err := services.QB.QueryRow("SELECT email FROM users WHERE id = ?", ctx.User.ID)
		if err != nil {
			c.Error(err)
		}
		tu.Simple(c, tu.GetSimpleByFilename("settings/password.html"), messages, map[string]interface{}{
			"email": s["email"],
		})
	}()

	if ok, _ := services.CSRF.Validate(ctx.User.ID, c.PostForm("csrf")); !ok {
		sessions.AddMessage(c, msg.ErrorMessage{lu.T(c, "Your session has expired. Please try redoing what you were trying to do.")})
		return
	}

	var password string
	services.DB.Get(&password, "SELECT password_md5 FROM users WHERE id = ? LIMIT 1", ctx.User.ID)

	if err := au.CompareHashPasswords(
		password,
		c.PostForm("currentpassword"),
	); err != nil {
		messages = append(messages, msg.ErrorMessage{lu.T(c, "Wrong password.")})
		return
	}

	uq := new(common.UpdateQuery)

	email := strings.ToLower(c.PostForm("email"))

	var currentEmail string
	services.DB.Get(&currentEmail, "SELECT email FROM users WHERE id = ?", ctx.User.ID)

	if email != "" && email != currentEmail {
		if services.DB.QueryRow("SELECT 1 FROM users WHERE email LIKE ?", email).
			Scan(new(int)) != sql.ErrNoRows {
			messages = append(messages, msg.ErrorMessage{lu.T(c, "This email is already in use!")})
			return
		}

		if !uu.ValidateEmail(email) {
			messages = append(messages, msg.ErrorMessage{lu.T(c, "Please use a valid email address.")})
			return
		}

		uq.Add("email", email)
	}

	if c.PostForm("newpassword") != "" {
		if s := au.ValidatePassword(c.PostForm("newpassword")); s != "" {
			messages = append(messages, msg.ErrorMessage{lu.T(c, s)})
			return
		}
		pw, err := au.GeneratePassword(c.PostForm("newpassword"))
		if err == nil {
			uq.Add("password_md5", pw)
		}
		sess := sessions.GetSession(c)
		sess.Set("pw", cryptography.MakeMD5(pw))
		sess.Save()
	}

	_, err := services.DB.Exec("UPDATE users SET "+uq.Fields()+" WHERE id = ? LIMIT 1", append(uq.Parameters, ctx.User.ID)...)
	if err != nil {
		c.Error(err)
	}

	services.DB.Exec("UPDATE users SET flags = flags & ~3 WHERE id = ? LIMIT 1", ctx.User.ID)

	messages = append(messages, msg.SuccessMessage{lu.T(c, "Your settings have been saved.")})
}
