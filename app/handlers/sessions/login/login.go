package login

import (
	"database/sql"
	"errors"
	"html/template"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slog"

	"github.com/amplitude/analytics-go/amplitude"
	s "github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/mileusna/useragent"
	"github.com/osuAkatsuki/akatsuki-api/common"
	eh "github.com/osuAkatsuki/hanayo/app/handlers/errors"
	"github.com/osuAkatsuki/hanayo/app/models"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	au "github.com/osuAkatsuki/hanayo/app/usecases/auth"
	"github.com/osuAkatsuki/hanayo/app/usecases/auth/cryptography"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	su "github.com/osuAkatsuki/hanayo/app/usecases/sessions"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
	uu "github.com/osuAkatsuki/hanayo/app/usecases/user"
	"github.com/osuAkatsuki/otp-service-client-go/client"
	"golang.org/x/crypto/bcrypt"
)

type SessionData struct {
	ID              int
	Username        string
	Password        string
	PasswordVersion int
	Country         string
	pRaw            int64
	Privileges      common.UserPrivileges
	Flags           uint
}

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

	var data SessionData
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
		slog.ErrorContext(c, err.Error())
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
		slog.Error("Error comparing passwords", "error", err.Error())
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

	userOtp, otpErr := services.OTP.GetUserOtp(data.ID)
	var notFoundErr *client.NotFoundError
	if otpErr != nil && !errors.As(otpErr, &notFoundErr) {
		slog.Error("Error checking user's OTP", "error", otpErr.Error())
		tu.SimpleReply(c, msg.ErrorMessage{lu.T(c, "Something went wrong. Please try again.")})
		return
	}

	rememberedDeviceId, err := c.Cookie("rd")
	rememberedDevice := err == nil && rememberedDeviceId != ""

	if rememberedDevice {
		_, rdErr := services.OTP.GetRememberedDevice(rememberedDeviceId)
		if rdErr != nil && !errors.As(otpErr, &notFoundErr) {
			slog.Error("Error checking user's remembered device", "error", rdErr.Error())
			tu.SimpleReply(c, msg.ErrorMessage{lu.T(c, "Something went wrong. Please try again.")})
			return
		}

		rememberedDevice = rdErr == nil
	}

	// if we managed to fetch an OTP, it's verified, enabled and not a remembered then we must use it
	if otpErr == nil && userOtp.Verified && userOtp.Enabled && !rememberedDevice {
		sess.Set("mustotp", true)
		sess.Set("user_id", data.ID)
		sess.Save()
		c.Redirect(302, "/login/otp")
		return
	}

	handleSession(c, sess, data)
}

func AfterLogin(c *gin.Context, id int, country string, flags uint) {
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

func LoginOtpPageHandler(c *gin.Context) {
	sess := sessions.GetSession(c)

	if sess.Get("user_id") == nil || !sess.Get("mustotp").(bool) {
		tu.Resp403(c)
		return
	}

	otpResponse(c)
}

const InvalidToken = "invalid_token"

func LoginOtpSubmitHandler(c *gin.Context) {
	sess := sessions.GetSession(c)

	if sess.Get("user_id") == nil || !sess.Get("mustotp").(bool) {
		tu.Resp403(c)
		return
	}

	token := c.PostForm("token")
	if token == "" {
		otpResponse(c, msg.ErrorMessage{lu.T(c, "You must enter a 2FA code.")})
		return
	}

	rememberDevice := c.PostForm("remember_device") == "on"
	userId := sess.Get("user_id").(int)

	err := services.OTP.ValidateOtp(userId, token)

	var badRequestErr *client.BadRequestError
	if err != nil && errors.As(err, &badRequestErr) && err.(*client.BadRequestError).Problem == InvalidToken {
		otpResponse(c, msg.ErrorMessage{lu.T(c, "Incorrect code.")})
		return
	} else if err != nil {
		otpResponse(c, msg.ErrorMessage{lu.T(c, "Something went wrong. Please try again.")})
		return
	}

	sess.Delete("user_id")
	sess.Delete("mustotp")

	var data SessionData
	err = services.DB.QueryRow(`
	SELECT
		u.id, u.username, u.password_md5,
		s.country, u.privileges, u.flags
	FROM users u
	LEFT JOIN users_stats s ON s.id = u.id
	WHERE u.id = ? LIMIT 1`, userId).Scan(
		&data.ID, &data.Username, &data.Password,
		&data.Country, &data.pRaw, &data.Flags,
	)
	data.Privileges = common.UserPrivileges(data.pRaw)
	slog.Info("got user", "user", data)

	if err != nil {
		slog.Error("Error fetching user for 2FA", "error", err.Error())
		otpResponse(c, msg.ErrorMessage{lu.T(c, "Something went wrong. Please try again.")})
		return
	}

	if rememberDevice {
		uu.SetRememberedDeviceCookie(data.ID, c)
	}

	handleSession(c, sess, data)
}

func handleSession(c *gin.Context, sess s.Session, data SessionData) {
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

func otpResponse(c *gin.Context, messages ...msg.Message) {
	sess := sessions.GetSession(c)

	if sess.Get("user_id") == nil || !sess.Get("mustotp").(bool) {
		tu.Resp403(c)
		return
	}

	tu.Resp(c, 200, "2fa_login.html", &models.BaseTemplateData{
		TitleBar:      lu.T(c, "2FA required for login"),
		BannerContent: "login2.jpg",
		BannerType:    1,
		Messages:      messages,
	})
}
