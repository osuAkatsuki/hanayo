package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"zxq.co/ripple/rippleapi/common"
	"zxq.co/x/rs"
)

func loginSubmit(c *gin.Context) {
	if getContext(c).User.ID != 0 {
		simpleReply(c, errorMessage{T(c, "You're already logged in!")})
		return
	}

	if c.PostForm("username") == "" || c.PostForm("password") == "" {
		simpleReply(c, errorMessage{T(c, "Username or password not set.")})
		return
	}

	param := "safe_name"
	u := c.PostForm("username")
	if strings.Contains(u, "@") {
		param = "email"
	} else {
		u = safeUsername(u)
	}

	var data struct {
		ID              int
		Username        string
		Password        string
		Country         string
		pRaw            int64
		Privileges      Privileges
	}
	err := db.QueryRow(`
	SELECT 
		u.id, u.pw_bcrypt, u.name,
		u.country, u.priv
	FROM users u
	WHERE u.`+param+` = ? LIMIT 1`, strings.TrimSpace(u)).Scan(
		&data.ID, &data.Password, &data.Username,
		&data.Country, &data.pRaw,
	)
	data.Privileges = Privileges(data.pRaw)

	switch {
	case err == sql.ErrNoRows:
		if param == "safe_name" {
			param = "name"
		}
		simpleReply(c, errorMessage{T(c, "No user with such %s!", param)})
		return
	case err != nil:
		c.Error(err)
		resp500(c)
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(data.Password),
		[]byte(cmd5(c.PostForm("password"))),
	); err != nil {
		simpleReply(c, errorMessage{T(c, "Wrong password.")})
		return
	}

	// update password if cost is < bcrypt.DefaultCost
	if i, err := bcrypt.Cost([]byte(data.Password)); err == nil && i < bcrypt.DefaultCost {
		pass, err := bcrypt.GenerateFromPassword([]byte(cmd5(c.PostForm("password"))), bcrypt.DefaultCost)
		if err == nil {
			if _, err := db.Exec("UPDATE users SET pw_bcrypt = ? WHERE id = ?", string(pass), data.ID); err == nil {
				data.Password = string(pass)
			}
		}
	}

	sess := getSession(c)

	if data.Privileges & VERIFIED == 0 {
		setYCookie(data.ID, c)
		addMessage(c, warningMessage{T(c, "You will need to verify your account first.")})
		sess.Save()
		c.Redirect(302, "/register/verify?u="+strconv.Itoa(data.ID))
		return
	}

	if data.Privileges & NORMAL == 0 {
		simpleReply(c, errorMessage{T(c, "You are not allowed to login. This means your account is either banned or locked.")})
		return
	}

	setYCookie(data.ID, c)

	sess.Set("userid", data.ID)
	sess.Set("pw", cmd5(data.Password))
	sess.Set("logout", rs.String(15))

	afterLogin(c, data.ID, data.Country)

	redir := c.PostForm("redir")
	if len(redir) > 0 && redir[0] != '/' {
		redir = ""
	}

	addMessage(c, successMessage{T(c, "Hey %s! You are now logged in.", template.HTMLEscapeString(data.Username))})
	sess.Save()
	if redir == "" {
		redir = "/"
	}
	c.Redirect(302, redir)
	return
}

func afterLogin(c *gin.Context, id int, country string) {
	s, err := generateToken(id, c)
	if err != nil {
		resp500(c)
		c.Error(err)
		return
	}
	getSession(c).Set("token", s)
	if country == "XX" {
		setCountry(c, id)
	}
	logIP(c, id)
}

func safeUsername(u string) string {
	return strings.Replace(strings.TrimSpace(strings.ToLower(u)), " ", "_", -1)
}

func logout(c *gin.Context) {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		respEmpty(c, "Log out", warningMessage{T(c, "You're already logged out!")})
		return
	}
	sess := getSession(c)
	s, _ := sess.Get("logout").(string)
	if s != c.Query("k") {
		// todo: return "are you sure you want to log out?" page
		respEmpty(c, "Log out", warningMessage{T(c, "Your session has expired. Please try redoing what you were trying to do.")})
		return
	}
	sess.Clear()
	http.SetCookie(c.Writer, &http.Cookie{
		Name:    "rt",
		Value:   "",
		Expires: time.Now().Add(-time.Hour),
	})
	addMessage(c, successMessage{T(c, "Successfully logged out.")})
	sess.Save()
	c.Redirect(302, "/")
}

func generateToken(id int, c *gin.Context) (string, error) {
	tok := common.RandomString(32)
	_, err := db.Exec(
		`INSERT INTO tokens(user, privileges, description, token, private)
					VALUES (   ?,        '0',           ?,     ?,     '1');`,
		id, clientIP(c), cmd5(tok))
	if err != nil {
		return "", err
	}
	return tok, nil
}

func checkToken(s string, id int, c *gin.Context) (string, error) {
	if s == "" {
		return generateToken(id, c)
	}
	if err := db.QueryRow("SELECT 1 FROM tokens WHERE token = ?", cmd5(s)).Scan(new(int)); err == sql.ErrNoRows {
		return generateToken(id, c)
	} else if err != nil {
		return "", err
	}
	return s, nil
}
