package main

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
)

func loadSimplePages(r *gin.Engine) {
	for _, el := range simplePages {
		if el.Handler == "" {
			continue
		}
		r.GET(el.Handler, simplePageFunc(el))
	}
}

func simplePageFunc(p templateConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := getContext(c)
		if s.User.Privileges&p.mp() != p.mp() {
			resp403(c)
			return
		}
		simple(c, p, nil, nil)
	}
}

func resp403(c *gin.Context) {
	if getContext(c).User.ID == 0 {
		ru := c.Request.URL
		addMessage(c, warningMessage{T(c, "You need to login first.")})
		getSession(c).Save()
		c.Redirect(302, "/login?redir="+url.QueryEscape(ru.Path+"?"+ru.RawQuery))
		return
	}
	respEmpty(c, "Forbidden", warningMessage{T(c, "You should not be 'round here.")})
}

func simpleReply(c *gin.Context, errs ...message) error {
	t := getSimple(c.Request.URL.Path)
	if t.Handler == "" {
		return errors.New("simpleReply: simplepage not found")
	}
	simple(c, t, errs, nil)
	return nil
}

func getSimple(h string) templateConfig {
	for _, s := range simplePages {
		if s.Handler == h {
			return s
		}
	}
	fmt.Println("oh handler shit not found", h)
	return templateConfig{}
}

func getSimpleByFilename(f string) templateConfig {
	for _, s := range simplePages {
		if s.Template == f {
			return s
		}
	}
	fmt.Println("oh shit not found", f)
	return templateConfig{}
}

func simple(c *gin.Context, p templateConfig, errs []message, requestInfo map[string]interface{}) {
	resp(c, 200, p.Template, &baseTemplateData{
		TitleBar:       p.TitleBar,
		KyutGrill:      p.KyutGrill,
		Scripts:        p.additionalJS(),
		HeadingOnRight: p.HugeHeadingRight,
		FormData:       normaliseURLValues(c.Request.PostForm),
		RequestInfo:    requestInfo,
		Messages:       errs,
	})
}

func normaliseURLValues(uv url.Values) map[string]string {
	m := make(map[string]string, len(uv))
	for k, v := range uv {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return m
}
