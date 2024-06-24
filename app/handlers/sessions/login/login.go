package login

import (
	"database/sql"
	"html/template"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slog"

	"github.com/amplitude/analytics-go/amplitude"
	"github.com/gin-gonic/gin"
	"github.com/mileusna/useragent"
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
		ID         int
		Username   string
		Password   string
		Country    string
		pRaw       int64
		Privileges common.UserPrivileges
	}
	err := services.DB.QueryRow(`
	SELECT
		id, password_md5,
		username, country, privileges
	FROM users
	WHERE `+param+` = ?`, strings.TrimSpace(u)).Scan(
		&data.ID, &data.Password,
		&data.Username,
		&data.Country, &data.pRaw,
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
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}

	if err := au.CompareHashPasswords(
		data.Password,
		c.PostForm("password"),
	); err != nil {
		slog.WarnContext(
			c,
			"User login failed due to incorrect credentials",
			"error",
			err.Error(),
			"input_username",
			data.Username,
		)
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
		} else {
			slog.Error("Error updating password", "error", err.Error())
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

	latitude, err := strconv.ParseFloat(c.Request.Header.Get("CF-IPLatitude"), 64)
	if err != nil {
		slog.Error("Error parsing latitude", "error", err.Error())
		latitude = 0.0
	}

	longitude, err := strconv.ParseFloat(c.Request.Header.Get("CF-IPLongitude"), 64)
	if err != nil {
		slog.Error("Error parsing longitude", "error", err.Error())
		longitude = 0.0
	}

	userAgent := useragent.Parse(c.Request.UserAgent())

	services.Amplitude.Track(amplitude.Event{
		EventType: "web_login",
		EventOptions: amplitude.EventOptions{
			UserID:      strconv.Itoa(data.ID),
			IP:          c.ClientIP(),
			Country:     c.Request.Header.Get("CF-IPCountry"),
			City:        c.Request.Header.Get("CF-IPCity"),
			LocationLat: latitude,
			LocationLng: longitude,
			Region:      c.Request.Header.Get("CF-Region"),
			Language:    c.Request.Header.Get("Accept-Language"),
			OSName:      userAgent.OS,
			OSVersion:   userAgent.OSVersion,
			DeviceModel: userAgent.Device,
		},
		EventProperties: map[string]interface{}{"source": "hanayo"},
	})

	identifyObj := amplitude.Identify{}
	identifyObj.Set("last_login", time.Now().Unix())
	identifyObj.Set("last_login_country", data.Country)
	identifyObj.Set("last_login_ip", c.ClientIP())
	identifyObj.Set("os_name", userAgent.OS)
	identifyObj.Set("os_version", userAgent.OSVersion)
	identifyObj.Set("device_model", userAgent.Device)

	services.Amplitude.Identify(
		identifyObj,
		amplitude.EventOptions{
			UserID: strconv.Itoa(data.ID),
		},
	)

	uu.SetYCookie(data.ID, c)

	sess.Set("userid", data.ID)
	sess.Set("pw", cryptography.MakeMD5(data.Password))
	sess.Set("logout", common.RandomString(15))

	AfterLogin(c, data.ID, data.Country)

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

func AfterLogin(c *gin.Context, id int, country string) {
	s, err := su.GenerateToken(id, c)
	if err != nil {
		eh.Resp500(c)
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		return
	}
	sessions.GetSession(c).Set("token", s)
	if country == "XX" {
		uu.SetCountry(c, id)
	}

	err = uu.LogIP(c, id)
	if err != nil {
		slog.Error("Error logging IP", "error", err.Error())
	}
}
