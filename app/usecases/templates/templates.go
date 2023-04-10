package templates

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/models"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	mt "github.com/osuAkatsuki/hanayo/app/models/templates"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/usecases/funcmap"
	loc "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	"github.com/osuAkatsuki/hanayo/app/usecases/misc"
	"github.com/rjeczalik/notify"
	"github.com/thehowl/conf"
)

var templates = make(map[string]*template.Template)
var baseTemplates = [...]string{
	"web/templates/base.html",
	"web/templates/navbar.html",
	"web/templates/simplepag.html",
}

func Reloader() error {
	c := make(chan notify.EventInfo, 1)
	if err := notify.Watch("./web/templates/...", c, notify.All); err != nil {
		return err
	}
	go func() {
		var last time.Time
		for ev := range c {
			if !strings.HasSuffix(ev.Path(), ".html") || time.Since(last) < time.Second*3 {
				continue
			}
			fmt.Println("Change detected! Refreshing templates")
			simplePages = []mt.TemplateConfig{}
			LoadTemplates("")
			last = time.Now()
		}
		defer notify.Stop(c)
	}()
	return nil
}

func LoadTemplates(subdir string) {
	ts, err := ioutil.ReadDir("web/templates" + subdir)
	if err != nil {
		panic(err)
	}

	for _, i := range ts {
		// if it's a directory, load recursively
		if i.IsDir() && i.Name() != ".." && i.Name() != "." {
			LoadTemplates(subdir + "/" + i.Name())
			continue
		}

		// ignore non-html files
		if !strings.HasSuffix(i.Name(), ".html") {
			continue
		}

		fullName := "web/templates" + subdir + "/" + i.Name()
		_c := parseConfig(fullName)
		var c mt.TemplateConfig
		if _c != nil {
			c = *_c
		}
		if c.NoCompile {
			continue
		}

		var files = c.Inc("web/templates" + subdir + "/")
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
		templates[inName+i.Name()] = template.Must(template.New(i.Name()).Funcs(funcmap.FuncMap).ParseFiles(
			append(files, baseTemplates[:]...)...,
		))

		if _c != nil {
			simplePages = append(simplePages, *_c)
		}
	}
}

func parseConfig(s string) *mt.TemplateConfig {
	f, err := os.Open(s)
	defer f.Close()
	if err != nil {
		return nil
	}
	i := bufio.NewScanner(f)
	var inConfig bool
	var buff string
	var t mt.TemplateConfig
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
			t.Template = strings.TrimPrefix(s, "web/templates/")
			return &t
		}
		if !inConfig {
			continue
		}
		buff += u + "\n"
	}
	return nil
}

func Resp(c *gin.Context, statusCode int, tpl string, data interface{}) {
	if c == nil {
		return
	}
	t := templates[tpl]
	if t == nil {
		c.String(500, "Template not found! Please tell this to a dev!")
		return
	}
	sess := sessions.GetSession(c)
	if corrected, ok := data.(mt.Page); ok {
		corrected.SetMessages(sessions.GetMessages(c))
		corrected.SetPath(c.Request.URL.Path)
		corrected.SetContext(sessions.GetContext(c))
		corrected.SetGinContext(c)
		corrected.SetSession(sess)
		data = corrected // correct the data structure.
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

func RespEmpty(c *gin.Context, title string, messages ...msg.Message) {
	Resp(c, 200, "empty.html", &models.BaseTemplateData{TitleBar: title, Messages: messages})
}

func Resp403(c *gin.Context) {
	if sessions.GetContext(c).User.ID == 0 {
		ru := c.Request.URL
		sessions.AddMessage(c, msg.WarningMessage{loc.T(c, "You need to login first.")})
		sessions.GetSession(c).Save()
		c.Redirect(302, "/login?redir="+url.QueryEscape(ru.Path+"?"+ru.RawQuery))
		return
	}
	RespEmpty(c, "Forbidden", msg.WarningMessage{loc.T(c, "You should not be 'round here.")})
}

func SimpleReply(c *gin.Context, errs ...msg.Message) error {
	t := GetSimple(c.Request.URL.Path)
	if t.Handler == "" {
		return errors.New("simpleReply: simplepage not found")
	}
	Simple(c, t, errs, nil)
	return nil
}

func Simple(c *gin.Context, p mt.TemplateConfig, errs []msg.Message, requestInfo map[string]interface{}) {
	Resp(c, 200, p.Template, &models.BaseTemplateData{
		TitleBar:       p.TitleBar,
		BannerContent:  p.BannerContent,
		BannerType:     p.BannerType,
		Scripts:        p.GetAdditionalJS(),
		HeadingOnRight: p.HugeHeadingRight,
		FormData:       misc.NormaliseURLValues(c.Request.PostForm),
		RequestInfo:    requestInfo,
		Messages:       errs,
	})
}
