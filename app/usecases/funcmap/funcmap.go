package funcmap

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slog"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/osuAkatsuki/akatsuki-api/common"
	"github.com/osuAkatsuki/hanayo/app/models"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	settingsState "github.com/osuAkatsuki/hanayo/app/states/settings"
	"github.com/osuAkatsuki/hanayo/app/usecases/geoloc"
	"github.com/osuAkatsuki/hanayo/app/usecases/misc"
	"github.com/osuAkatsuki/hanayo/app/version"
	"github.com/osuAkatsuki/hanayo/internal/bbcode"
	"github.com/osuAkatsuki/hanayo/internal/doc"
	fasuimappings "github.com/osuAkatsuki/hanayo/internal/fa-semantic-mappings"
	"github.com/russross/blackfriday/v2"
	"github.com/thehowl/qsql"
	"zxq.co/ripple/playstyle"
)

const PREMIUM_PRICE_PER_MONTH = 5.0

// funcMap contains useful functions for the various templates.
var FuncMap = template.FuncMap{
	"print": fmt.Println,
	// html disables HTML escaping on the values it is given.
	"html": func(value interface{}) template.HTML {
		return template.HTML(fmt.Sprint(value))
	},
	// avatars is a function returning the configuration constant AvatarURL
	"config": func(key string) interface{} {
		return services.ConfigMap[key]
	},
	// navbarItem is a function to generate an item in the navbar.
	// The reason why this exists is that I wanted to have the currently
	// selected element in the navbar having the "active" class.
	"navbarItem": func(currentPath, name, path string) template.HTML {
		var act string
		if path == currentPath {
			act = "active "
		}
		return template.HTML(fmt.Sprintf(`<a class="%sitem" href="%s">%s</a>`, act, path, name))
	},
	// curryear returns the current year.
	"curryear": func() int {
		return time.Now().Year()
	},
	// hasAdmin returns, based on the user's privileges, whether they should be
	// able to see the RAP button (aka AdminPrivilegeAccessRAP).
	"hasAdmin": func(privs common.UserPrivileges) bool {
		return privs&common.AdminPrivilegeAccessRAP > 0
	},
	// isRAP returns whether the current page is in RAP.
	"isRAP": func(p string) bool {
		parts := strings.Split(p, "/")
		return len(parts) > 1 && parts[1] == "admin"
	},
	// favMode is just a helper function for user profiles. Basically checks
	// whether a float and an int are ==, and if they are it will return "active ",
	// so that the element in the mode menu of a user profile can be marked as
	// the current active element.
	"favMode": func(favMode float64, current int) string {
		if int(favMode) == current {
			return "active "
		}
		return ""
	},
	// slice generates a []interface{} with the elements it is given.
	// useful to iterate over some elements, like this:
	//  {{ range slice 1 2 3 }}{{ . }}{{ end }}
	"slice": func(els ...interface{}) []interface{} {
		return els
	},
	// int converts a float/int to an int.
	"int": func(f interface{}) int {
		if f == nil {
			return 0
		}
		switch f := f.(type) {
		case int:
			return f
		case float64:
			return int(f)
		case float32:
			return int(f)
		}
		return 0
	},
	// float converts an int to a float.
	"float": func(i int) float64 {
		return float64(i)
	},
	// atoi converts a string to an int and then a float64.
	// If s is not an actual int, it returns nil.
	"atoi": func(s string) interface{} {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil
		}
		return float64(i)
	},
	// atoint is like atoi but returns always an int.
	"atoint": func(s string) int {
		i, _ := strconv.Atoi(s)
		return i
	},
	// parseUserpage compiles BBCode to HTML.
	"parseUserpage": func(s string) template.HTML {
		return template.HTML(bbcode.ConvertBBCodeToHTML(s))
	},
	// time converts a RFC3339 timestamp to the HTML element <time>.
	"time": func(s string) template.HTML {
		t, _ := time.Parse(time.RFC3339, s)
		return _time(s, t)
	},
	// time generates a time from a native Go time.Time
	"timeFromTime": func(t time.Time) template.HTML {
		return _time(t.Format(time.RFC3339), t)
	},
	// timeAddDay is basically time but adds a day.
	"timeAddDay": func(s string) template.HTML {
		t, _ := time.Parse(time.RFC3339, s)
		t = t.Add(time.Hour * 24)
		return _time(t.Format(time.RFC3339), t)
	},
	// nativeTime creates a native Go time.Time from a RFC3339 timestamp.
	"nativeTime": func(s string) time.Time {
		t, _ := time.Parse(time.RFC3339, s)
		return t
	},
	"playtimeConv": func(input float64) string {
		/* Thanks to Night(#1429) for this! */

		// years := math.Floor(float64(input) / 60 / 60 / 24 / 7 / 30 / 12)
		// seconds := int(input) % (60 * 60 * 24 * 7 * 30 * 12)
		// months := math.Floor(float64(seconds) / 60 / 60 / 24 / 7 / 30)
		// seconds = int(input) % (60 * 60 * 24 * 7 * 30)
		// weeks := math.Floor(float64(seconds) / 60 / 60 / 24 / 7)
		// seconds = int(input) % (60 * 60 * 24 * 7)
		days := math.Floor(float64(input) / 60 / 60 / 24)
		seconds := int(input) % (60 * 60 * 24)
		hours := math.Floor(float64(seconds) / 60 / 60)
		seconds = int(input) % (60 * 60)
		minutes := math.Floor(float64(seconds) / 60)
		seconds = int(input) % 60

		return fmt.Sprintf("%01dd %01dh %01dm", int(math.Floor(days)), int(math.Floor(hours)), int(math.Floor(minutes)))
	},
	// band is a bitwise AND.
	"band": func(i1 int, i ...int) int {
		for _, el := range i {
			i1 &= el
		}
		return i1
	},
	// countryReadable converts a country's ISO name to its full name.
	"countryReadable":   geoloc.CountryReadable,
	"countryCodepoints": geoloc.CountryToCodepoints,
	"country": func(s string, name bool) template.HTML {
		var c string
		if name {
			c = geoloc.CountryReadable(s)
			if c == "" {
				return ""
			}
		}
		codepoints := geoloc.CountryToCodepoints(s)
		return template.HTML(fmt.Sprintf(`<img src="/static/images/flags/%s.svg" class="new-flag"></img>%s`, codepoints, c))
	},
	// humanize pretty-prints a float, e.g.
	//     humanize(1000) == "1,000"
	"humanize": func(f float64) string {
		return humanize.Commaf(f)
	},
	// levelPercent basically does this:
	//     levelPercent(56.23215) == "23"
	"levelPercent": func(l float64) string {
		_, f := math.Modf(l)
		f *= 100
		return fmt.Sprintf("%.0f", f)
	},
	// level removes the decimal part from a float.
	"level": func(l float64) string {
		i, _ := math.Modf(l)
		return fmt.Sprintf("%.0f", i)
	},
	// faIcon converts a fontawesome icon to a semantic ui icon.
	"faIcon": func(i string) string {
		classes := strings.Split(i, " ")
		for i, class := range classes {
			if v, ok := fasuimappings.Mappings[class]; ok {
				classes[i] = v
			}
		}
		return strings.Join(classes, " ")
	},
	// log fmt.Printf's something
	"log": fmt.Printf,
	// has returns whether priv1 has all 1 bits of priv2, aka priv1 & priv2 == priv2
	"has": func(priv1 interface{}, priv2 float64) bool {
		var p1 uint64
		switch priv1 := priv1.(type) {
		case common.UserPrivileges:
			p1 = uint64(priv1)
		case float64:
			p1 = uint64(priv1)
		case int:
			p1 = uint64(priv1)
		}
		return p1&uint64(priv2) == uint64(priv2)
	},
	// _range is like python range's.
	// If it is given 1 argument, it returns a []int containing numbers from 0
	// to x.
	// If it is given 2 arguments, it returns a []int containing numers from x
	// to y if x < y, from y to x if y < x.
	"_range": func(x int, y ...int) ([]int, error) {
		switch len(y) {
		case 0:
			r := make([]int, x)
			for i := range r {
				r[i] = i
			}
			return r, nil
		case 1:
			nums, up := pos(y[0] - x)
			r := make([]int, nums)
			for i := range r {
				if up {
					r[i] = i + x + 1
				} else {
					r[i] = i + y[0]
				}
			}
			if !up {
				// reverse r
				sort.Sort(sort.Reverse(sort.IntSlice(r)))
			}
			return r, nil
		}
		return nil, errors.New("y must be at maximum 1 parameter")
	},
	// blackfriday passes some markdown through blackfriday.
	"blackfriday": func(m string) template.HTML {
		// The reason of m[strings.Index...] is to remove the "header", where
		// there is the information about the file (namely, title, old_id and
		// reference_version)
		return template.HTML(
			blackfriday.Run(
				[]byte(
					m[strings.Index(m, "\n---\n")+5:],
				),
				blackfriday.WithExtensions(blackfriday.CommonExtensions),
			),
		)
	},
	// i is an inline if.
	// i (cond) (true) (false)
	"i": func(a bool, x, y interface{}) interface{} {
		if a {
			return x
		}
		return y
	},
	// modes returns an array containing all the modes (in their string representation).
	"modes": func() []string {
		return []string{
			"osu!",
			"Taiko",
			"Catch",
			"Mania",
		}
	},

	"cmodes": func() []string {
		return []string{
			"Vanilla",
			"Relax",
			"Autopilot",
		}
	},
	// _or is like or, but has only false and nil as its "falsey" values
	"_or": func(args ...interface{}) interface{} {
		for _, a := range args {
			if a != nil && a != false {
				return a
			}
		}
		return nil
	},
	// unixNano returns the UNIX timestamp of when hanayo was started in nanoseconds.
	"unixNano": func() string {
		return strconv.FormatInt(hanayoStarted, 10)
	},
	// playstyle returns the string representation of a playstyle.
	"playstyle": func(i float64, f *models.ProfileData) string {
		var parts []string

		p := int(i)
		for k, v := range playstyle.Styles {
			if p&(1<<uint(k)) > 0 {
				parts = append(parts, f.T(v))
			}
		}

		return strings.Join(parts, ", ")
	},
	// arithmetic plus/minus
	"plus": func(i ...float64) float64 {
		var sum float64
		for _, i := range i {
			sum += i
		}
		return sum
	},
	"minus": func(i1 float64, i ...float64) float64 {
		for _, i := range i {
			i1 -= i
		}
		return i1
	},
	// rsin - Return Slice If Nil
	"rsin": func(i interface{}) interface{} {
		if i == nil {
			return []struct{}{}
		}
		return i
	},
	// loadjson loads a json file.
	"loadjson": func(jsonfile string) interface{} {
		f, err := ioutil.ReadFile(jsonfile)
		if err != nil {
			return nil
		}
		var x interface{}
		err = json.Unmarshal(f, &x)
		if err != nil {
			return nil
		}
		return x
	},
	// teamJSON returns the data of team.json
	"teamJSON": func() map[string]interface{} {
		f, err := ioutil.ReadFile("team.json")
		if err != nil {
			return nil
		}
		var m map[string]interface{}
		json.Unmarshal(f, &m)
		return m
	},
	// in returns whether the first argument is in one of the following
	"in": func(a1 interface{}, as ...interface{}) bool {
		for _, a := range as {
			if a == a1 {
				return true
			}
		}
		return false
	},
	"capitalise": strings.Title,
	// servicePrefix gets the prefix of a service, like github.
	"servicePrefix": func(s string) string { return servicePrefixes[s] },
	// randomLogoColour picks a "random" colour for ripple's logo.
	"randomLogoColour": func() string {
		if rand.Int()%4 == 0 {
			return logoColours[rand.Int()%len(logoColours)]
		}
		return "pink"
	},
	// after checks whether a certain time is after time.Now()
	"after": func(s string) bool {
		t, _ := time.Parse(time.RFC3339, s)
		return t.After(time.Now())
	},

	// qsql functions
	"qb": func(q string, p ...interface{}) map[string]qsql.String {
		r, err := services.QB.QueryRow(q, p...)
		if err != nil {
			slog.Error("Error in qb", "error", err.Error())
		}
		if r == nil {
			return make(map[string]qsql.String, 0)
		}
		return r
	},
	"qba": func(q string, p ...interface{}) []map[string]qsql.String {
		r, err := services.QB.Query(q, p...)
		if err != nil {
			slog.Error("Error in qba", "error", err.Error())
		}
		return r
	},
	"qbe": func(q string, p ...interface{}) int {
		i, _, err := services.QB.Exec(q, p...)
		if err != nil {
			slog.Error("Error in qbe", "error", err.Error())
		}
		return i
	},

	// bget makes a request to the bancho api
	// https://docs.ripple.moe/docs/banchoapi/v1
	"bget": func(ept string, qs ...interface{}) map[string]interface{} {
		settings := settingsState.GetSettings()
		d, err := http.Get(fmt.Sprintf(settings.PUBLIC_BANCHO_SERVICE_BASE_URL+"/api/v1/"+ept, qs...))
		if err != nil {
			slog.Error("Error in bget", "error", err.Error())
			return nil
		}
		x := make(map[string]interface{})
		data, _ := ioutil.ReadAll(d.Body)
		json.Unmarshal(data, &x)
		return x
	},
	// styles returns playstyle.Styles
	"styles": func() []string {
		return playstyle.Styles[:]
	},
	// shift shifts n1 by n2
	"shift": func(n1, n2 int) int {
		return n1 << uint(n2)
	},
	// calculatePremiumPrice calculates the price of x premium months in euros.
	"calculatePremiumPrice": func(months float64) string {
		return fmt.Sprintf("%.2f", months * PREMIUM_PRICE_PER_MONTH)
	},
	// csrfGenerate creates a csrf token input
	"csrfGenerate": func(u int) template.HTML {
		return template.HTML(`<input type="hidden" name="csrf" value="` + misc.MustCSRFGenerate(u) + `">`)
	},
	// csrfURL creates a CSRF token for GET requests.
	"csrfURL": func(u int) template.URL {
		return template.URL("csrf=" + misc.MustCSRFGenerate(u))
	},
	// systemSetting retrieves some information from the table system_settings
	"systemSettings": systemSettings,
	// authCodeURL gets the auth code for discord
	"authCodeURL": func(u int) string {
		return "" //getDiscord().AuthCodeURL(misc.MustCSRFGenerate(u))
	},
	// perc returns a percentage
	"perc": func(i, total float64) string {
		return fmt.Sprintf("%.0f", i/total*100)
	},
	// atLeastOne returns 1 if i < 1, or i otherwise.
	"atLeastOne": func(i int) int {
		if i < 1 {
			i = 1
		}
		return i
	},
	// ieForm fixes forms in IE/Trident being immensely fucked up. I hate microsoft.
	"ieForm": func(c *gin.Context) template.HTML {
		if !IsIE(c.Request.UserAgent()) {
			return ""
		}
		return ieUnfucker
	},
	// version gets what's the current Hanayo version.
	"version": func() string {
		return version.Version
	},
	// "generateKey": generateKey,
	// rediget retrieves a value from redis.
	"rediget": func(k string) string {
		x := services.RD.Get(k)
		if x == nil {
			return ""
		}
		if err := x.Err(); err != nil {
			slog.Error("An error occurred", "error", err.Error())
		}
		return x.Val()
	},
	"languageInformation": func() []langInfo {
		return languageInformation
	},
	"languageInformationByNameShort": func(s string) langInfo {
		for _, lang := range languageInformation {
			if lang.NameShort == s {
				return lang
			}
		}
		return langInfo{}
	},
	"countryList": func(n int64) []string {
		return services.RD.ZRevRange("hanayo:country_list", 0, n-1).Val()
	},
	"documentationFiles": doc.GetDocs,
	"documentationData": func(slug string, language string) doc.File {
		if i, err := strconv.Atoi(slug); err == nil {
			slug = doc.SlugFromOldID(i)
		}
		return doc.GetFile(slug, language)
	},
	"privilegesToString": func(privs float64) string {
		return common.Privileges(privs).String()
	},
	"htmlescaper": template.HTMLEscaper,
}

var hanayoStarted = time.Now().UnixNano()

var servicePrefixes = map[string]string{
	"github":  "https://github.com/",
	"twitter": "https://twitter.com/",
	"mail":    "mailto:",
}

var logoColours = [...]string{
	"blue",
	"green",
	"orange",
	"red",
}

// we still haven't got jquery when the script is here, so well shit.
const ieUnfucker = `<input type="submit" class="ie" name="submit" value="submit">
<script>
var deferredToPageLoad = function() {
	$("button[form]").on("click", function() {
		$("form#" + $(this).attr("form") + " input.ie").click();
	});
};
</script>`

func pos(x int) (int, bool) {
	if x > 0 {
		return x, true
	}
	return x * -1, false
}
func _time(s string, t time.Time) template.HTML {
	return template.HTML(fmt.Sprintf(`<time class="timeago" datetime="%s">%v</time>`, s, t))
}

// Fantastic IEs And Where To Find Them
var ieUserAgentsContain = []string{
	"MSIE ",
	"Trident/",
	"Edge/",
}

func IsIE(s string) bool {
	for _, v := range ieUserAgentsContain {
		if strings.Contains(s, v) {
			return true
		}
	}
	return false
}

type systemSetting struct {
	Name   string
	Int    int
	String string
}

func systemSettings(names ...string) map[string]systemSetting {
	var settingsRaw []systemSetting
	q, p, _ := sqlx.In("SELECT name, value_int as `int`, value_string as `string` FROM system_settings WHERE name IN (?)", names)
	err := services.DB.Select(&settingsRaw, q, p...)
	if err != nil {
		slog.Error("Error in systemSettings", err.Error())
		return nil
	}
	settings := make(map[string]systemSetting, len(names))
	for _, s := range settingsRaw {
		settings[s.Name] = s
	}
	return settings
}

// func getDiscord() *oauth2.Config {
// 	return &oauth2.Config{
// 		ClientID:     settings.Config.DiscordOAuthID,
// 		ClientSecret: settings.Config.DiscordOAuthSecret,
// 		RedirectURL:  settings.Config.BaseURL + "/settings/discord/finish",
// 		Endpoint:     discordoauth.Endpoint,
// 		Scopes:       []string{"identify"},
// 	}
// }

func init() {
	rand.Seed(time.Now().UnixNano())
}

type langInfo struct {
	Name, CountryShort, NameShort string
}

var languageInformation = []langInfo{
	{"Deutsch", "de", "de"},
	{"English (UK)", "gb", "en"},
	{"Español", "es", "es"},
	{"Français", "fr", "fr"},
	{"Italiano", "it", "it"},
	{"Nederlands", "nl", "nl"},
	{"Polski", "pl", "pl"},
	{"Русский", "ru", "ru"},
	{"Română", "ro", "ro"},
	{"Suomi", "fi", "fi"},
	{"Svenska", "se", "sv"},
	{"Tiếng Việt Nam", "vn", "vi"},
	{"한국어", "kr", "ko"},
}
