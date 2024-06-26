package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	settingsState "github.com/osuAkatsuki/hanayo/app/states/settings"
	beatmaps "github.com/osuAkatsuki/hanayo/app/usecases/beatmaps"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
)

type BaseTemplateData struct {
	TitleBar       string // required
	HeadingTitle   string
	HeadingOnRight bool
	Scripts        []string
	BannerContent  string
	BannerAbsolute bool
	BannerType     int  // 0 = none, 1 = image, 2 = colour, 3 = animated.
	DisableHH      bool // HH = Huge Heading
	Messages       []msg.Message
	RequestInfo    map[string]interface{}

	// ignore, they're set by resp()
	Context  Context
	Path     string
	FormData map[string]string
	Gin      *gin.Context
	Session  sessions.Session
}

func (b *BaseTemplateData) SetMessages(m []msg.Message) {
	b.Messages = append(b.Messages, m...)
}
func (b *BaseTemplateData) SetPath(path string) {
	b.Path = path
}
func (b *BaseTemplateData) SetContext(c Context) {
	b.Context = c
}
func (b *BaseTemplateData) SetGinContext(c *gin.Context) {
	b.Gin = c
}
func (b *BaseTemplateData) SetSession(sess sessions.Session) {
	b.Session = sess
}
func (b BaseTemplateData) Get(s string, params ...interface{}) map[string]interface{} {
	s = fmt.Sprintf(s, params...)
	settings := settingsState.GetSettings()
	req, err := http.NewRequest("GET", settings.INTERNAL_AKATSUKI_API_BASE_URL+"/"+s, nil)
	if err != nil {
		b.Gin.Error(err)
		return nil
	}
	req.Header.Set("User-Agent", "hanayo")
	req.Header.Set("H-Key", settings.APP_HANAYO_KEY)
	req.Header.Set("X-Ripple-Token", b.Context.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		b.Gin.Error(err)
		return nil
	}
	data, err := io.ReadAll(resp.Body)
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
func (b BaseTemplateData) Has(privs uint64) bool {
	return uint64(b.Context.User.Privileges)&privs == privs
}
func (b BaseTemplateData) Conf() interface{} {
	settings := settingsState.GetSettings()
	return settings
}

func (b *BaseTemplateData) T(s string, args ...interface{}) string {
	return lu.T(b.Gin, s, args...)
}

type BeatmapPageData struct {
	BaseTemplateData

	Found      bool
	Beatmap    beatmaps.Beatmap
	Beatmapset beatmaps.BeatmapSet
	SetJSON    string
}

type MatchData struct {
	BaseTemplateData
	MatchID   int
	MatchName string
}

type ClanData struct {
	BaseTemplateData
	ClanID int
}

type ProfileData struct {
	BaseTemplateData
	UserID int
}
