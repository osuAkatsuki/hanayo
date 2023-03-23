package templates

import (
	"strings"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/akatsuki-api/common"
	"github.com/osuAkatsuki/hanayo/app/models"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
)

type TemplateConfig struct {
	NoCompile bool
	Include   string
	Template  string

	// Stuff that used to be in simpleTemplate
	Handler          string
	TitleBar         string
	BannerContent    string
	BannerType       int
	MinPrivileges    uint64
	HugeHeadingRight bool
	AdditionalJS     string
}

func (t TemplateConfig) Inc(prefix string) []string {
	if t.Include == "" {
		return nil
	}
	a := strings.Split(t.Include, ",")
	for i, s := range a {
		a[i] = prefix + s
	}
	return a
}

func (t TemplateConfig) MP() common.UserPrivileges {
	return common.UserPrivileges(t.MinPrivileges)
}

func (t TemplateConfig) GetAdditionalJS() []string {
	parts := strings.Split(t.AdditionalJS, ",")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

type Page interface {
	SetMessages([]msg.Message)
	SetPath(string)
	SetContext(models.Context)
	SetGinContext(*gin.Context)
	SetSession(sessions.Session)
}
