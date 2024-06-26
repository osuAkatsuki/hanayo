package register

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slog"

	"github.com/amplitude/analytics-go/amplitude"
	"github.com/gin-gonic/gin"
	"github.com/mileusna/useragent"
	"github.com/osuAkatsuki/akatsuki-api/common"
	eh "github.com/osuAkatsuki/hanayo/app/handlers/errors"
	"github.com/osuAkatsuki/hanayo/app/models"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	settingsState "github.com/osuAkatsuki/hanayo/app/states/settings"
	au "github.com/osuAkatsuki/hanayo/app/usecases/auth"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	"github.com/osuAkatsuki/hanayo/app/usecases/misc"
	su "github.com/osuAkatsuki/hanayo/app/usecases/sessions"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
	uu "github.com/osuAkatsuki/hanayo/app/usecases/user"
)

func RegisterPageHandler(c *gin.Context) {
	if sessions.GetContext(c).User.ID != 0 {
		tu.Resp403(c)
		return
	}
	if c.Query("stopsign") != "1" {
		u, _ := tryBotnets(c)
		if u != "" {
			tu.Simple(c, tu.GetSimpleByFilename("register/elmo.html"), nil, map[string]interface{}{
				"Username": u,
			})
			return
		}
	}
	registerResp(c)
}

func RegisterSubmitHandler(c *gin.Context) {
	settings := settingsState.GetSettings()
	if sessions.GetContext(c).User.ID != 0 {
		tu.Resp403(c)
		return
	}
	// check registrations are enabled
	if !registrationsEnabled() {
		registerResp(c, msg.ErrorMessage{lu.T(c, "Sorry, it's not possible to register at the moment. Please try again later.")})
		return
	}

	// check username is valid by our criteria
	username := strings.TrimSpace(c.PostForm("username"))
	email := strings.ToLower(c.PostForm("email"))

	if x := uu.ValidateUsername(username); x != "" {
		registerResp(c, msg.ErrorMessage{lu.T(c, x)})
		return
	}

	/* beta keys
	key := strings.TrimSpace(c.PostForm("key"))
	if db.QueryRow("SELECT 1 FROM beta_keys WHERE key = ?", c.PostForm("key")).
		Scan(new(int)) ==  sql.ErrNoRows {
		registerResp(c, errorMessage{T(c, "Your key is invalid!")})
		return
	}
	*/

	// check if key is required for login and if passed
	if keyRequired() && c.PostForm("key") == "" {
		registerResp(c, msg.ErrorMessage{lu.T(c, "Please pass a valid key.")})
		return
	}
	// check if given key is valid
	if !checkKey(c.PostForm("key")) && keyRequired() {
		registerResp(c, msg.ErrorMessage{lu.T(c, "Please pass a valid key.")})
		return
	}

	// check email is valid
	if !uu.ValidateEmail(email) {
		registerResp(c, msg.ErrorMessage{lu.T(c, "Please use a valid email address.")})
		return
	}

	if c.PostForm("password") != c.PostForm("password2") {
		registerResp(c, msg.ErrorMessage{lu.T(c, "The passwords doesn't match!")})
		return
	}

	// passwords check (too short/too common)
	if x := au.ValidatePassword(c.PostForm("password")); x != "" {
		registerResp(c, msg.ErrorMessage{lu.T(c, x)})
		return
	}

	// usernames with both _ and spaces are not allowed
	if strings.Contains(username, "_") && strings.Contains(username, " ") {
		registerResp(c, msg.ErrorMessage{lu.T(c, "An username can't contain both underscores and spaces.")})
		return
	}

	// check whether username already exists
	if services.DB.QueryRow("SELECT 1 FROM users WHERE username_safe = ?", uu.SafeUsername(username)).
		Scan(new(int)) != sql.ErrNoRows {
		registerResp(c, msg.ErrorMessage{lu.T(c, "An user with that username already exists!")})
		return
	}

	// check whether an user with that email already exists
	if services.DB.QueryRow("SELECT 1 FROM users WHERE email LIKE ?", email).
		Scan(new(int)) != sql.ErrNoRows {
		registerResp(c, msg.ErrorMessage{lu.T(c, "An user with that email address already exists!")})
		return
	}

	// recaptcha verify
	if settings.RECAPTCHA_SECRET_KEY != "" && !misc.RecaptchaCheck(c) {
		registerResp(c, msg.ErrorMessage{lu.T(c, "Captcha is invalid.")})
		return
	}

	// TODO: make it send discord webhook
	// uMulti, criteria := tryBotnets(c)
	// if criteria != "" {
	// 		fmt.Sprintf(
	// 			"User **%s** registered with the same %s as %s (%s/u/%s). **POSSIBLE MULTIACCOUNT!!!**. Waiting for ingame verification...",
	// 			username, criteria, uMulti, settings.APP_BASE_URL, url.QueryEscape(uMulti),
	// 		),
	// }

	// The actual registration.
	pass, err := au.GeneratePassword(c.PostForm("password"))
	if err != nil {
		slog.Error("Error generating password", "error", err.Error())
		eh.Resp500(c)
		return
	}

	if len(c.Request.Header.Get("CF-IPCountry")) > 2 {
		registerResp(c, msg.ErrorMessage{lu.T(c, "Cloudflare error.")})
		return
	}

	tx, err := services.DB.Begin()
	if err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}

	res, err := tx.Exec(`INSERT INTO users(username, username_safe, password_md5, email, register_datetime, privileges, latest_activity) VALUES (?, ?, ?, ?, ?, ?, ?);`,
		username, uu.SafeUsername(username), pass, email, time.Now().Unix(), common.UserPrivilegePendingVerification, time.Now().Unix())

	if err != nil {
		tx.Rollback()
		c.Error(err)
		slog.ErrorContext(c, err.Error())
		eh.Resp500(c)
		return
	}

	userId, _ := res.LastInsertId()

	for _, mode := range []int{0, 1, 2, 3, 4, 5, 6, 8} {
		_, err = tx.Exec("INSERT INTO `user_stats` (user_id, mode) VALUES (?, ?);", userId, mode)
		if err != nil {
			tx.Rollback()
			c.Error(err)
			slog.ErrorContext(c, err.Error())
			eh.Resp500(c)
			return
		}
	}

	tx.Commit()

	/* Beta Keys
	db.Exec("UPDATE `beta_keys` set used = 1 where key = ?", key)

	// Ripple Gay Bot
	schiavo.CMs.Send(fmt.Sprintf("User (**%s** | %s) registered from %s", username, c.PostForm("email"), clientIP(c)))
	*/
	// delete the key c
	//db.Exec("DELETE FROM beta_keys WHERE beta_key = ?", c.PostForm("key"))

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
		EventType: "web_signup",
		EventOptions: amplitude.EventOptions{
			UserID:      strconv.FormatInt(userId, 10),
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
	identifyObj.SetOnce("username", username)
	identifyObj.SetOnce("email", email)
	identifyObj.SetOnce("signup_date", time.Now().Unix())
	identifyObj.SetOnce("signup_ip", c.ClientIP())

	services.Amplitude.Identify(
		identifyObj,
		amplitude.EventOptions{
			UserID: strconv.FormatInt(userId, 10),
		},
	)

	uu.SetYCookie(int(userId), c)

	err = uu.LogIP(c, int(userId))
	if err != nil {
		slog.Error("Error logging IP", "error", err.Error())
	}

	services.RD.Incr("ripple:registered_users")

	//addMessage(c, successMessage{T(c, "You have been successfully registered on Akatsuki!")})
	sessions.GetSession(c).Save()
	c.Redirect(302, "/register/verify?u="+strconv.Itoa(int(userId)))
}

func registerResp(c *gin.Context, messages ...msg.Message) {
	tu.Resp(c, 200, "register/register.html", &models.BaseTemplateData{
		TitleBar:      "Register",
		BannerContent: "register.jpg",
		BannerType:    1,
		Scripts:       []string{"https://www.google.com/recaptcha/api.js"},
		Messages:      messages,
		FormData:      misc.NormaliseURLValues(c.Request.PostForm),
	})
}

func registrationsEnabled() bool {
	var enabled bool
	services.DB.QueryRow("SELECT value_int FROM system_settings WHERE name = 'registrations_enabled'").Scan(&enabled)
	return enabled
}

func keyRequired() bool {
	var enabled bool
	services.DB.QueryRow("SELECT value_int FROM system_settings WHERE name = 'regkey_required'").Scan(&enabled)
	return enabled
}

func checkKey(passed string) bool {
	if services.DB.QueryRow("SELECT beta_key FROM beta_keys WHERE beta_key = ?", passed).Scan(new(int)) != sql.ErrNoRows {
		return true
	} else {
		return false
	}
}

func VerifyAccountPageHandler(c *gin.Context) {
	if sessions.GetContext(c).User.ID != 0 {
		tu.Resp403(c)
		return
	}

	/*
		i, ret := checkUInQS(c)
		if ret {
			return
		}

		sess := getSession(c)
		var rPrivileges uint64
		db.Get(&rPrivileges, "SELECT privileges FROM users WHERE id = ?", i)
		if common.UserPrivileges(rPrivileges)&common.UserPrivilegePendingVerification == 0 {
			//addMessage(c, warningMessage{T(c, "Nope.")})
			sess.Save()
			//c.Redirect(302, "/")
			//return
		}*/

	sessions.AddMessage(c, msg.SuccessMessage{lu.T(c, "You have been successfully registered on Akatsuki!")})

	tu.Resp(c, 200, "register/verify.html", &models.BaseTemplateData{
		TitleBar:       "Welcome to Akatsuki!",
		HeadingOnRight: true,
		BannerContent:  "welcome.jpg",
		BannerType:     1,
	})
}

func WelcomePageHandler(c *gin.Context) {
	if sessions.GetContext(c).User.ID != 0 {
		tu.Resp403(c)
		return
	}

	i, ret := checkUInQS(c)
	if ret {
		return
	}

	var rPrivileges uint64
	services.DB.Get(&rPrivileges, "SELECT privileges FROM users WHERE id = ?", i)
	if common.UserPrivileges(rPrivileges)&common.UserPrivilegePendingVerification > 0 {
		c.Redirect(302, "/register/verify?u="+c.Query("u"))
		return
	}

	t := lu.T(c, "Welcome!")
	if common.UserPrivileges(rPrivileges)&common.UserPrivilegeNormal == 0 {
		// if the user has no UserNormal, it means they're banned = they multiaccounted
		t = lu.T(c, "Welcome back!")
	}

	tu.Resp(c, 200, "register/welcome.html", &models.BaseTemplateData{
		TitleBar:       t,
		HeadingOnRight: true,
		BannerContent:  "welcome.jpg",
		BannerType:     1,
	})
}

// Check User In Query Is Same As User In Y Cookie
func checkUInQS(c *gin.Context) (int, bool) {
	sess := sessions.GetSession(c)

	i, _ := strconv.Atoi(c.Query("u"))
	y, _ := c.Cookie("y")
	err := services.DB.QueryRow("SELECT 1 FROM identity_tokens WHERE token = ? AND userid = ?", y, i).Scan(new(int))
	if err == sql.ErrNoRows {
		sessions.AddMessage(c, msg.WarningMessage{lu.T(c, "Nope.")})
		sess.Save()
		c.Redirect(302, "/")
		return 0, true
	}
	return i, false
}

func tryBotnets(c *gin.Context) (string, string) {
	var username string

	err := services.DB.QueryRow("SELECT u.username FROM ip_user i INNER JOIN users u ON u.id = i.userid WHERE i.ip = ?", su.ClientIP(c)).Scan(&username)
	if err != nil {
		if err != sql.ErrNoRows {
			c.Error(err)
			slog.ErrorContext(c, err.Error())
		}
		return "", ""
	}
	if username != "" {
		return username, "IP"
	}

	cook, _ := c.Cookie("y")
	err = services.DB.QueryRow("SELECT u.username FROM identity_tokens i INNER JOIN users u ON u.id = i.userid WHERE i.token = ?",
		cook).Scan(&username)
	if err != nil {
		if err != sql.ErrNoRows {
			c.Error(err)
			slog.ErrorContext(c, err.Error())
		}
		return "", ""
	}
	if username != "" {
		return username, "username"
	}

	return "", ""
}
