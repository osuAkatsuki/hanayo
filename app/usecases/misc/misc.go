package misc

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	settingsState "github.com/osuAkatsuki/hanayo/app/states/settings"
	su "github.com/osuAkatsuki/hanayo/app/usecases/sessions"
)

//go:generate go run scripts/generate_mappings.go -g
//go:generate go run scripts/top_passwords.go

func RecaptchaCheck(c *gin.Context) bool {
	settings := settingsState.GetSettings()
	f := make(url.Values)
	f.Add("secret", settings.RECAPTCHA_SECRET_KEY)
	f.Add("response", c.PostForm("g-recaptcha-response"))
	f.Add("remoteip", su.ClientIP(c))

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

func NormaliseURLValues(uv url.Values) map[string]string {
	m := make(map[string]string, len(uv))
	for k, v := range uv {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return m
}

func MustCSRFGenerate(u int) string {
	v, err := services.CSRF.Generate(u)
	if err != nil {
		panic(err)
	}
	return v
}
