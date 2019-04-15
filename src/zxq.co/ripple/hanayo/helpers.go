package main

import (
	gocontext "context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bytes"

	"github.com/gin-gonic/gin"
	"zxq.co/ripple/hanayo/modules/bbcode"
	tp "zxq.co/ripple/hanayo/modules/top-passwords"
	"zxq.co/ripple/rippleapi/common"
)

//go:generate go run scripts/generate_mappings.go -g
//go:generate go run scripts/top_passwords.go

func cmd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func validatePassword(p string) string {
	if len(p) < 8 {
		return "Your password is too short! It must be at least 8 characters long."
	}

	for _, k := range tp.TopPasswords {
		if k == p {
			return "Your password is one of the most common passwords on the entire internet. No way we're letting you use that!"
		}
	}

	return ""
}

func recaptchaCheck(c *gin.Context) bool {
	f := make(url.Values)
	f.Add("secret", config.RecaptchaPrivate)
	f.Add("response", c.PostForm("g-recaptcha-response"))
	f.Add("remoteip", clientIP(c))

	req, err := http.Post("https://www.google.com/recaptcha/api/siteverify",
		"application/x-www-form-urlencoded", strings.NewReader(f.Encode()))
	if err != nil {
		c.Error(err)
		return false
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		c.Error(err)
		return false
	}

	var e struct {
		Success bool `json:"success"`
	}
	err = json.Unmarshal(data, &e)
	if err != nil {
		c.Error(err)
		return false
	}

	return e.Success
}

func parseBBCode(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.Error(err)
		c.String(200, "Error")
		return
	}
	d := bbcode.Compile(string(body))
	c.String(200, d)
}

func discordFinish(c *gin.Context) {
	sess := getSession(c)
	defer func() {
		sess.Save()
		c.Redirect(302, "/settings/discord")
	}()

	ctx := getContext(c)
	if ok, _ := CSRF.Validate(ctx.User.ID, c.Query("state")); !ok {
		addMessage(c, errorMessage{T(c, "Your session has expired. Please try redoing what you were trying to do.")})
		return
	}

	if ctx.User.Privileges&common.UserPrivilegeDonor == 0 {
		addMessage(c, errorMessage{T(c, "You're not a donor!")})
		return
	}

	reqCtx, _ := gocontext.WithTimeout(gocontext.Background(), time.Second*20)
	tok, err := getDiscord().Exchange(reqCtx, c.Query("code"))
	if err != nil {
		c.Error(err)
		addMessage(c, errorMessage{T(c, "An error occurred.")})
		return
	}

	// Yoloest error handling ever
	// Here we're getting the user ID of our user on discord
	req, _ := http.NewRequest("GET", "https://discordapp.com/api/users/@me", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	resp, _ := http.DefaultClient.Do(req)
	rawData, _ := ioutil.ReadAll(resp.Body)
	var x struct {
		ID string `json:"id"`
	}
	err = json.Unmarshal(rawData, &x)
	if err != nil {
		c.Error(err)
		addMessage(c, errorMessage{T(c, "An error occurred.")})
		return
	}

	// Here, instead, we're telling donorbot about the user.
	// setup post data
	vals := make(url.Values, 2)
	vals.Set("discord_id", x.ID)
	vals.Set("secret", config.DonorBotSecret)
	// send request
	resp, err = http.Post(config.DonorBotURL+"/api/v1/give_donor", "application/x-www-form-urlencoded", bytes.NewReader([]byte(vals.Encode())))
	if err != nil {
		c.Error(err)
		addMessage(c, errorMessage{T(c, "An error occurred.")})
		return
	}

	var o struct {
		Status int `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&o)

	switch o.Status {
	case 200:
		// move on
	case 404:
		addMessage(c, errorMessage{T(c, "You've not joined the discord server! Links to it are below on the page. Please join the server before attempting to connect your account to Discord.")})
	default:
		c.Error(fmt.Errorf("donorbot: %d", resp.StatusCode))
		addMessage(c, errorMessage{T(c, "An error occurred.")})
		return
	}

	db.Exec("INSERT INTO discord_roles (id, userid, discordid, roleid) VALUES (NULL, ?, ?, 0)", ctx.User.ID, x.ID)

	addMessage(c, successMessage{T(c, "Your account has been linked successfully!")})
}

func mustCSRFGenerate(u int) string {
	v, err := CSRF.Generate(u)
	if err != nil {
		panic(err)
	}
	return v
}

var blogRedirectMap = map[string]string{
	"/posts/moving-to-a-new-server":                                              "https://blog.ripple.moe/moving-to-a-new-server-11155949edca",
	"/posts/ripple-qa-3":                                                         "https://blog.ripple.moe/ripple-q-a-3-28c9851f42b3",
	"/posts/hanayo-is-now-the-ripple-website":                                    "https://blog.ripple.moe/hanayo-is-now-the-ripple-website-3bcfaab60c4f",
	"/posts/ripple-qa-2":                                                         "https://blog.ripple.moe/ripple-q-a-2-1204be5ffeef",
	"/posts/hanayo-is-live":                                                      "https://blog.ripple.moe/hanayo-is-live-still-not-replacing-the-official-site-f2b751a5baf7",
	"/posts/ripple-qa-1":                                                         "https://blog.ripple.moe/ripple-q-a-1-51181dd8df65",
	"/posts/why-am-i-randomly-gaining-losing-pp":                                 "https://blog.ripple.moe/why-am-i-randomly-gaining-losing-pp-595aedfdc5db",
	"/posts/more-love-for-donors":                                                "https://blog.ripple.moe/more-love-for-donors-96c889a9d95f",
	"/posts/happy-birthday-ripple":                                               "https://blog.ripple.moe/happy-birthday-ripple-fbc4bbc47936",
	"/posts/going-back-open-source":                                              "https://blog.ripple.moe/going-back-open-source-a53469e15658",
	"/posts/the-useless-things-we-make-during-weekends-series-continues":         "https://blog.ripple.moe/the-useless-things-we-make-during-weekends-series-continues-1a06671ff5c2",
	"/posts/performance-points-pp":                                               "https://blog.ripple.moe/performance-points-pp-d02e0353ad81",
	"/posts/why-are-you-introducing-so-many-bugs-its-not-like-we-asked-for-them": "https://blog.ripple.moe/why-are-you-introducing-so-many-bugs-its-not-like-we-asked-for-them-c650a8ea9667",
	"/posts/going-closed-source":                                                 "https://blog.ripple.moe/going-closed-source-5c0a991f581f",
	"/posts/changes-in-administration":                                           "https://blog.ripple.moe/changes-in-administration-983114dc6332",
	"/posts/its-dangerous-to-go-alone":                                           "https://blog.ripple.moe/its-dangerous-to-go-alone-ef7fa98f2975",
	"/posts/we-got-a-blog":                                                       "https://blog.ripple.moe/we-got-a-blog-81a0af62b410",
}

func blogRedirect(c *gin.Context) {
	a := c.Param("url")
	red := blogRedirectMap[a]
	if red == "" {
		red = "https://blog.ripple.moe"
	}
	c.Redirect(301, red)
}
