package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pariz/gountries"
	"github.com/rjeczalik/notify"
	"github.com/thehowl/conf"
	"zxq.co/ripple/rippleapi/common"
)

var templates = make(map[string]*template.Template)
var baseTemplates = [...]string{
	"templates/base.html",
	"templates/navbar.html",
	"templates/simplepag.html",
}
var simplePages []templateConfig

var gdb = gountries.New()

func countryReadable(s string) string {
	if s == "XX" || s == "" {
		return ""
	}
	reg, err := gdb.FindCountryByAlpha(s)
	if err != nil {
		return ""
	}
	return reg.Name.Common
}

func loadTemplates(subdir string) {
	ts, err := ioutil.ReadDir("templates" + subdir)
	if err != nil {
		panic(err)
	}

	for _, i := range ts {
		// if it's a directory, load recursively
		if i.IsDir() && i.Name() != ".." && i.Name() != "." {
			loadTemplates(subdir + "/" + i.Name())
			continue
		}

		// ignore non-html files
		if strings.HasPrefix(i.Name(), ".html") {
			continue
		}

		fullName := "templates" + subdir + "/" + i.Name()
		_c := parseConfig(fullName)
		var c templateConfig
		if _c != nil {
			c = *_c
		}
		if c.NoCompile {
			continue
		}

		var files = c.inc("templates" + subdir + "/")
		files = append(files, fullName)

		// do not compile base templates on their own
		var comp bool
		for _, j := range baseTemplates {
			if fullName == j {
				comp = true
				break
			}
		}
		if comp {
			continue
		}

		var inName string
		if subdir != "" && subdir[0] == '/' {
			inName = subdir[1:] + "/"
		}

		// add new template to template slice
		templates[inName+i.Name()] = template.Must(template.New(i.Name()).Funcs(funcMap).ParseFiles(
			append(files, baseTemplates[:]...)...,
		))

		if _c != nil {
			simplePages = append(simplePages, *_c)
		}
	}
}

func resp(c *gin.Context, statusCode int, tpl string, data interface{}) {
	if c == nil {
		return
	}
	t := templates[tpl]
	if t == nil {
		c.String(500, "Template not found! Please tell this to a dev!")
		return
	}
	sess := getSession(c)
	if corrected, ok := data.(page); ok {
		corrected.SetMessages(getMessages(c))
		corrected.SetPath(c.Request.URL.Path)
		corrected.SetContext(getContext(c))
		corrected.SetGinContext(c)
		corrected.SetSession(sess)
	}
	sess.Save()
	buf := &bytes.Buffer{}
	err := t.ExecuteTemplate(buf, "base", data)
	if err != nil {
		c.String(
			200,
			"An error occurred while trying to render the page, and we have now been notified about it.",
		)
		c.Error(err)
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(statusCode)
	_, err = io.Copy(c.Writer, buf)
	if err != nil {
		c.Writer.WriteString("We don't know what's happening now.")
		c.Error(err)
		return
	}
}

type baseTemplateData struct {
	TitleBar          string // required
	HeadingTitle      string
	HeadingOnRight    bool
	Scripts           []string
	KyutGrill         string
	KyutGrillAbsolute bool
	SolidColour       string
	DisableHH         bool // HH = Huge Heading
	Messages          []message
	RequestInfo       map[string]interface{}

	// ignore, they're set by resp()
	Context  context
	Path     string
	FormData map[string]string
	Gin      *gin.Context
	Session  sessions.Session
}

func (b *baseTemplateData) SetMessages(m []message) {
	b.Messages = append(b.Messages, m...)
}
func (b *baseTemplateData) SetPath(path string) {
	b.Path = path
}
func (b *baseTemplateData) SetContext(c context) {
	b.Context = c
}
func (b *baseTemplateData) SetGinContext(c *gin.Context) {
	b.Gin = c
}
func (b *baseTemplateData) SetSession(sess sessions.Session) {
	b.Session = sess
}
func (b baseTemplateData) Get(s string, params ...interface{}) map[string]interface{} {
	s = fmt.Sprintf(s, params...)
	req, err := http.NewRequest("GET", config.API+s, nil)
	if err != nil {
		b.Gin.Error(err)
		return nil
	}
	req.Header.Set("User-Agent", "hanayo")
	req.Header.Set("H-Key", config.APISecret)
	req.Header.Set("X-Ripple-Token", b.Context.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		b.Gin.Error(err)
		return nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		b.Gin.Error(err)
		return nil
	}
	x := make(map[string]interface{})
	err = json.Unmarshal(data, &x)
	if err != nil {
		b.Gin.Error(err)
		return nil
	}
	return x
}
func (b baseTemplateData) Has(privs uint64) bool {
	return uint64(b.Context.User.Privileges)&privs == privs
}
func (b baseTemplateData) Conf() interface{} {
	return config
}

// list of client flags
const (
	CFDarkSite = 1 << iota
)

func (b baseTemplateData) ClientFlags() int {
	s, _ := b.Gin.Cookie("cflags")
	return common.Int(s)
}

type page interface {
	SetMessages([]message)
	SetPath(string)
	SetContext(context)
	SetGinContext(*gin.Context)
	SetSession(sessions.Session)
}

func reloader() error {
	c := make(chan notify.EventInfo, 1)
	if err := notify.Watch("./templates/...", c, notify.All); err != nil {
		return err
	}
	go func() {
		var last time.Time
		for ev := range c {
			if !strings.HasSuffix(ev.Path(), ".html") || time.Since(last) < time.Second*3 {
				continue
			}
			fmt.Println("Change detected! Refreshing templates")
			simplePages = []templateConfig{}
			loadTemplates("")
			l.Close()
			last = time.Now()
		}
		defer notify.Stop(c)
	}()
	return nil
}

type templateConfig struct {
	NoCompile bool
	Include   string
	Template  string

	// Stuff that used to be in simpleTemplate
	Handler          string
	TitleBar         string
	KyutGrill        string
	MinPrivileges    uint64
	HugeHeadingRight bool
	AdditionalJS     string
}

func (t templateConfig) inc(prefix string) []string {
	if t.Include == "" {
		return nil
	}
	a := strings.Split(t.Include, ",")
	for i, s := range a {
		a[i] = prefix + s
	}
	return a
}

func (t templateConfig) mp() common.UserPrivileges {
	return common.UserPrivileges(t.MinPrivileges)
}

func (t templateConfig) additionalJS() []string {
	parts := strings.Split(t.AdditionalJS, ",")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func parseConfig(s string) *templateConfig {
	f, err := os.Open(s)
	defer f.Close()
	if err != nil {
		return nil
	}
	i := bufio.NewScanner(f)
	var inConfig bool
	var buff string
	var t templateConfig
	for i.Scan() {
		u := i.Text()
		switch u {
		case "{{/*###":
			inConfig = true
		case "*/}}":
			if !inConfig {
				continue
			}
			conf.LoadRaw(&t, []byte(buff))
			t.Template = strings.TrimPrefix(s, "templates/")
			return &t
		}
		if !inConfig {
			continue
		}
		buff += u + "\n"
	}
	return nil
}

func respEmpty(c *gin.Context, title string, messages ...message) {
	resp(c, 200, "empty.html", &baseTemplateData{TitleBar: title, Messages: messages})
}
