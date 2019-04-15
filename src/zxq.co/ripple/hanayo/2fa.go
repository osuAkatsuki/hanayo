package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"zxq.co/ripple/rippleapi/common"
	"zxq.co/x/rs"
)

var allowedPaths = [...]string{
	"/logout",
	"/2fa_gateway",
	"/2fa_gateway/verify",
	"/2fa_gateway/clear",
	"/2fa_gateway/recover",
	"/favicon.ico",
}

// middleware to deny all requests to non-allowed pages
func twoFALock(c *gin.Context) {
	// check it's not a static file
	if len(c.Request.URL.Path) >= 8 && c.Request.URL.Path[:8] == "/static/" {
		c.Next()
		return
	}

	ctx := getContext(c)
	if ctx.User.ID == 0 {
		c.Next()
		return
	}

	sess := getSession(c)
	if v, _ := sess.Get("2fa_must_validate").(bool); !v {
		// * check 2fa is enabled.
		//   if it is,
		//   * check whether the current ip is found in the database.
		//     if it is, move on and show the page.
		//     if it isn't, set 2fa_must_validate
		//   if it isn't, move on.
		if is2faEnabled(ctx.User.ID) > 0 {
			err := db.QueryRow("SELECT 1 FROM ip_user WHERE userid = ? AND ip = ? LIMIT 1", ctx.User.ID, clientIP(c)).Scan(new(int))
			if err != sql.ErrNoRows {
				c.Next()
				return
			}
			sess.Set("2fa_must_validate", true)
		} else {
			c.Next()
			return
		}
	}

	// check it's one of the few approved paths
	for _, a := range allowedPaths {
		if a == c.Request.URL.Path {
			sess.Save()
			c.Next()
			return
		}
	}
	addMessage(c, warningMessage{T(c, "You need to complete the 2fa challenge first.")})
	sess.Save()
	query := c.Request.URL.RawQuery
	if query != "" {
		query = "?" + query
	}
	c.Redirect(302, "/2fa_gateway?redir="+url.QueryEscape(c.Request.URL.Path+query))
	c.Abort()
}

const (
	tfaEnabledTelegram = 1 << iota
	tfaEnabledTOTP
)

// is2faEnabled checks 2fa is enabled for an user.
func is2faEnabled(user int) int {
	var enabled int
	db.QueryRow("SELECT IFNULL((SELECT 1 FROM 2fa_telegram WHERE userid = ?), 0) | IFNULL((SELECT 2 FROM 2fa_totp WHERE userid = ? AND enabled = 1), 0) as x", user, user).
		Scan(&enabled)
	return enabled
}

func tfaGateway(c *gin.Context) {
	sess := getSession(c)

	redir := c.Query("redir")
	switch {
	case redir == "":
		redir = "/"
	case redir[0] != '/':
		redir = "/"
	}

	i, _ := sess.Get("userid").(int)
	if i == 0 {
		c.Redirect(302, redir)
	}

	// check 2fa hasn't been disabled
	e := is2faEnabled(i)
	if e == 0 {
		sess.Delete("2fa_must_validate")
		sess.Save()
		c.Redirect(302, redir)
		return
	}

	if e == 1 {
		// check previous 2fa thing is still valid
		err := db.QueryRow("SELECT 1 FROM 2fa WHERE userid = ? AND ip = ? AND expire > ?",
			i, clientIP(c), time.Now().Unix()).Scan(new(int))
		if err != nil {
			db.Exec("INSERT INTO 2fa(userid, token, ip, expire, sent) VALUES (?, ?, ?, ?, 0);",
				i, strings.ToUpper(rs.String(8)), clientIP(c), time.Now().Add(time.Hour).Unix())
			http.Get("http://127.0.0.1:8888/update")
		}
	}

	resp(c, 200, "2fa_gateway.html", &baseTemplateData{
		TitleBar:  "Two Factor Authentication",
		KyutGrill: "2fa.jpg",
		RequestInfo: map[string]interface{}{
			"redir": redir,
		},
	})
}

func clientIP(c *gin.Context) string {
	ff := c.Request.Header.Get("CF-Connecting-IP")
	if ff != "" {
		return ff
	}
	return c.ClientIP()
}

func clear2fa(c *gin.Context) {
	// basically deletes from db 2fa tokens, so that it gets regenerated when user hits gateway page
	sess := getSession(c)
	i, _ := sess.Get("userid").(int)
	if i == 0 {
		c.Redirect(302, "/")
	}
	db.Exec("DELETE FROM 2fa WHERE userid = ? AND ip = ?", i, clientIP(c))
	addMessage(c, successMessage{T(c, "A new code has been generated and sent to you through Telegram.")})
	sess.Save()
	c.Redirect(302, "/2fa_gateway")
}

func verify2fa(c *gin.Context) {
	sess := getSession(c)
	i, _ := sess.Get("userid").(int)
	if i == 0 {
		c.Redirect(302, "/")
	}
	e := is2faEnabled(i)
	switch e {
	case 1:
		var id int
		var expire common.UnixTimestamp
		err := db.QueryRow("SELECT id, expire FROM 2fa WHERE userid = ? AND ip = ? AND token = ?", i, clientIP(c), strings.ToUpper(c.Query("token"))).Scan(&id, &expire)
		if err == sql.ErrNoRows {
			c.String(200, "1")
			return
		}
		if time.Now().After(time.Time(expire)) {
			c.String(200, "1")
			db.Exec("INSERT INTO 2fa(userid, token, ip, expire, sent) VALUES (?, ?, ?, ?, 0);",
				i, strings.ToUpper(rs.String(8)), clientIP(c), time.Now().Add(time.Hour).Unix())
			http.Get("http://127.0.0.1:8888/update")
			return
		}
	case 2:
		var secret string
		db.Get(&secret, "SELECT secret FROM 2fa_totp WHERE userid = ?", i)
		if !totp.Validate(strings.Replace(c.Query("token"), " ", "", -1), secret) {
			c.String(200, "1")
			return
		}
	}

	loginUser(c, i)

	db.Exec("DELETE FROM 2fa WHERE id = ?", i)
	c.String(200, "0")
}

func loginUser(c *gin.Context, i int) {
	var d struct {
		Country string
		Flags   uint
	}
	err := db.Get(&d, "SELECT users_stats.country, users.flags FROM users_stats "+
		"LEFT JOIN users ON users.id = users_stats.id WHERE users_stats.id = ?", i)
	if err != nil {
		c.Error(err)
	}

	afterLogin(c, i, d.Country, d.Flags)

	addMessage(c, successMessage{T(c, "You've been successfully logged in.")})

	sess := getSession(c)
	sess.Delete("2fa_must_validate")
	sess.Save()
}

func recover2fa(c *gin.Context) {
	sess := getSession(c)
	i, _ := sess.Get("userid").(int)
	if i == 0 {
		c.Redirect(302, "/")
	}
	e := is2faEnabled(i)
	if e != 2 {
		respEmpty(c, "Recover account", warningMessage{T(c, "Oh no you don't.")})
		return
	}
	resp(c, 200, "2fa_gateway_recover.html", &baseTemplateData{
		TitleBar:  T(c, "Recover account"),
		KyutGrill: "2fa.jpg",
	})
}

func recover2faSubmit(c *gin.Context) {
	sess := getSession(c)
	i, _ := sess.Get("userid").(int)
	if i == 0 {
		c.Redirect(302, "/")
	}
	if is2faEnabled(i) != 2 {
		respEmpty(c, T(c, "Recover account"), warningMessage{T(c, "Get out.")})
		return
	}

	var codesRaw string
	db.Get(&codesRaw, "SELECT recovery FROM 2fa_totp WHERE userid = ?", i)
	var codes []string
	json.Unmarshal([]byte(codesRaw), &codes)

	for k, v := range codes {
		if v == c.PostForm("recovery_code") {
			codes[k] = codes[len(codes)-1]
			codes = codes[:len(codes)-1]
			b, _ := json.Marshal(codes)
			db.Exec("UPDATE 2fa_totp SET recovery = ? WHERE userid = ?", string(b), i)

			loginUser(c, i)
			c.Redirect(302, "/")
			return
		}
	}

	resp(c, 200, "2fa_gateway_recover.html", &baseTemplateData{
		TitleBar:  T(c, "Recover account"),
		KyutGrill: "2fa.jpg",
		Messages:  []message{errorMessage{T(c, "Recovery code is invalid.")}},
	})
}

// deletes expired 2fa confirmation tokens. gets current confirmation token.
// if it does not exist, generates one.
func get2faConfirmationToken(user int) (token string) {
	db.Exec("DELETE FROM 2fa_confirmation WHERE expire < ?", time.Now().Unix())
	db.Get(&token, "SELECT token FROM 2fa_confirmation WHERE userid = ? LIMIT 1", user)
	if token != "" {
		return
	}
	token = rs.String(32)
	db.Exec("INSERT INTO 2fa_confirmation (userid, token, expire) VALUES (?, ?, ?)",
		user, token, time.Now().Add(time.Hour).Unix())
	return
}

func disable2fa(c *gin.Context) {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		resp403(c)
		return
	}

	s := getSession(c)
	var m message
	defer func() {
		addMessage(c, m)
		s.Save()
		c.Redirect(302, "/settings/2fa")
	}()

	if ok, _ := CSRF.Validate(ctx.User.ID, c.PostForm("csrf")); !ok {
		m = errorMessage{T(c, "Your session has expired. Please try redoing what you were trying to do.")}
		return
	}

	var pass string
	db.Get(&pass, "SELECT password_md5 FROM users WHERE id = ?", ctx.User.ID)
	if err := bcrypt.CompareHashAndPassword(
		[]byte(pass),
		[]byte(cmd5(c.PostForm("password"))),
	); err != nil {
		m = errorMessage{"Wrong password."}
		return
	}

	db.Exec("DELETE FROM 2fa_telegram WHERE userid = ?", ctx.User.ID)
	db.Exec("DELETE FROM 2fa_totp WHERE userid = ?", ctx.User.ID)
	m = successMessage{T(c, "2FA disabled successfully.")}
}

func totpSetup(c *gin.Context) {
	ctx := getContext(c)
	sess := getSession(c)
	if ctx.User.ID == 0 {
		resp403(c)
		return
	}
	defer c.Redirect(302, "/settings/2fa")
	defer sess.Save()

	if ok, _ := CSRF.Validate(ctx.User.ID, c.PostForm("csrf")); !ok {
		addMessage(c, errorMessage{T(c, "Your session has expired. Please try redoing what you were trying to do.")})
		return
	}

	switch is2faEnabled(ctx.User.ID) {
	case 1:
		addMessage(c, errorMessage{T(c, "You currently have Telegram 2FA enabled. You first need to disable that if you want to use TOTP-based 2FA.")})
		return
	case 2:
		addMessage(c, errorMessage{T(c, "TOTP-based 2FA is already enabled!")})
		return
	}

	pc := strings.Replace(c.PostForm("passcode"), " ", "", -1)

	var secret string
	db.Get(&secret, "SELECT secret FROM 2fa_totp WHERE userid = ?", ctx.User.ID)
	if secret == "" || pc == "" {
		addMessage(c, errorMessage{T(c, "No passcode/secret was given. Please try again")})
		return
	}

	fmt.Println(pc, secret)
	if !totp.Validate(pc, secret) {
		addMessage(c, errorMessage{T(c, "Passcode is invalid. Perhaps it expired?")})
		return
	}

	codes, _ := json.Marshal(generateRecoveryCodes())
	db.Exec("UPDATE 2fa_totp SET recovery = ?, enabled = 1 WHERE userid = ?", string(codes), ctx.User.ID)

	addMessage(c, successMessage{T(c, "TOTP-based 2FA has been enabled on your account.")})
}

func generateRecoveryCodes() []string {
	x := make([]string, 8)
	for i := range x {
		x[i] = rs.StringFromChars(6, "QWERTYUIOPASDFGHJKLZXCVBNM1234567890")
	}
	return x
}

func generateKey(ctx context) *otp.Key {
	k, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Ripple",
		AccountName: ctx.User.Username,
	})
	if err != nil {
		return nil
	}
	db.Exec("INSERT INTO 2fa_totp(userid, secret) VALUES (?, ?) ON DUPLICATE KEY UPDATE secret = VALUES(secret)", ctx.User.ID, k.Secret())
	return k
}
