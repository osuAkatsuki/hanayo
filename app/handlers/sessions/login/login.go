package login

import (
	"database/sql"
	"html/template"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/akatsuki-api/common"
	eh "github.com/osuAkatsuki/hanayo/app/handlers/errors"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	au "github.com/osuAkatsuki/hanayo/app/usecases/auth"
	"github.com/osuAkatsuki/hanayo/app/usecases/auth/cryptography"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	su "github.com/osuAkatsuki/hanayo/app/usecases/sessions"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
	uu "github.com/osuAkatsuki/hanayo/app/usecases/user"
	"golang.org/x/crypto/bcrypt"
)

func LoginSubmitHandler(c *gin.Context) {
	if sessions.GetContext(c).User.ID != 0 {
		tu.SimpleReply(c, msg.ErrorMessage{lu.T(c, "You're already logged in!")})
		return
	}

	if c.PostForm("username") == "" || c.PostForm("password") == "" {
		tu.SimpleReply(c, msg.ErrorMessage{lu.T(c, "Username or password not set.")})
		return
	}

	param := "username_safe"
	u := c.PostForm("username")
	if strings.Contains(u, "@") {
		param = "email"
	} else {
		u = uu.SafeUsername(u)
	}

	var data struct {
		ID              int
		Username        string
		Password        string
		PasswordVersion int
		Country         string
		pRaw            int64
		Privileges      common.UserPrivileges
		Flags           uint
	}
	err := services.DB.QueryRow(`
	SELECT 
		u.id, u.password_md5,
		u.username, u.password_version,
		s.country, u.privileges, u.flags
	FROM users u
	LEFT JOIN users_stats s ON s.id = u.id
	WHERE u.`+param+` = ? LIMIT 1`, strings.TrimSpace(u)).Scan(
		&data.ID, &data.Password,
		&data.Username, &data.PasswordVersion,
		&data.Country, &data.pRaw, &data.Flags,
	)
	data.Privileges = common.UserPrivileges(data.pRaw)

	switch {
	case err == sql.ErrNoRows:
		if param == "username_safe" {
			param = "username"
		}
		tu.SimpleReply(c, msg.ErrorMessage{lu.T(c, "No user with such %s!", param)})
		return
	case err != nil:
		c.Error(err)
		eh.Resp500(c)
		return
	}

	if data.PasswordVersion == 1 {
		sessions.AddMessage(c, msg.WarningMessage{lu.T(c, "Your password is sooooooo old, that we don't even know how to deal with it anymore. Could you please change it?")})
		c.Redirect(302, "/pwreset")
		return
	}

	if err := au.CompareHashPasswords(
		data.Password,
		c.PostForm("password"),
	); err != nil {
		tu.SimpleReply(c, msg.ErrorMessage{lu.T(c, "Wrong password.")})
		return
	}

	// update password if cost is < bcrypt.DefaultCost
	if i, err := bcrypt.Cost([]byte(data.Password)); err == nil && i < bcrypt.DefaultCost {
		pass, err := au.GeneratePassword(c.PostForm("password"))
		if err == nil {
			if _, err := services.DB.Exec("UPDATE users SET password_md5 = ? WHERE id = ?", string(pass), data.ID); err == nil {
				data.Password = string(pass)
			}
		}
	}

	sess := sessions.GetSession(c)

	if data.Privileges&common.UserPrivilegePendingVerification > 0 {
		uu.SetYCookie(data.ID, c)
		sessions.AddMessage(c, msg.WarningMessage{lu.T(c, "You will need to verify your account first.")})
		sess.Save()
		c.Redirect(302, "/register/verify?u="+strconv.Itoa(data.ID))
		return
	}

	if data.Privileges&common.UserPrivilegeNormal == 0 {
		tu.SimpleReply(c, msg.ErrorMessage{lu.T(c, "You are not allowed to login. This means your account is either banned or locked.")})
		return
	}

	uu.SetYCookie(data.ID, c)

	sess.Set("userid", data.ID)
	sess.Set("pw", cryptography.MakeMD5(data.Password))
	sess.Set("logout", common.RandomString(15))

	AfterLogin(c, data.ID, data.Country, data.Flags)

	redir := c.PostForm("redir")
	if len(redir) > 0 && redir[0] != '/' {
		redir = ""
	}

	sessions.AddMessage(c, msg.SuccessMessage{lu.T(c, "Hey %s! You are now logged in.", template.HTMLEscapeString(data.Username))})
	sess.Save()
	if redir == "" {
		redir = "/"
	}
	c.Redirect(302, redir)
	return
}

func AfterLogin(c *gin.Context, id int, country string, flags uint) {
	s, err := su.GenerateToken(id, c)
	if err != nil {
		eh.Resp500(c)
		c.Error(err)
		return
	}
	sessions.GetSession(c).Set("token", s)
	if country == "XX" {
		uu.SetCountry(c, id)
	}
	uu.LogIP(c, id)
}
