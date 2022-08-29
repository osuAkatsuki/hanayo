package templates

import (
	"fmt"

	"github.com/gin-gonic/gin"
	mt "github.com/osuAkatsuki/hanayo/app/models/templates"
	"github.com/osuAkatsuki/hanayo/app/sessions"
)

var simplePages []mt.TemplateConfig

func LoadSimplePages(r *gin.Engine) {
	for _, el := range simplePages {
		if el.Handler == "" {
			continue
		}
		r.GET(el.Handler, simplePageFunc(el))
	}
}

func simplePageFunc(p mt.TemplateConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.GetContext(c)
		if s.User.Privileges&p.MP() != p.MP() {
			Resp403(c)
			return
		}
		Simple(c, p, nil, nil)
	}
}

func GetSimple(h string) mt.TemplateConfig {
	for _, s := range simplePages {
		if s.Handler == h {
			return s
		}
	}
	fmt.Println("oh handler shit not found", h)
	return mt.TemplateConfig{}
}

func GetSimpleByFilename(f string) mt.TemplateConfig {
	for _, s := range simplePages {
		if s.Template == f {
			return s
		}
	}
	fmt.Println("oh shit not found", f)
	return mt.TemplateConfig{}
}
