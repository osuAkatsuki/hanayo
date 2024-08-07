package user

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/slog"

	"github.com/gin-gonic/gin"
	"github.com/go-email-validator/go-email-validator/pkg/ev"
	"github.com/go-email-validator/go-email-validator/pkg/ev/contains"
	"github.com/go-email-validator/go-email-validator/pkg/ev/disposable"
	"github.com/go-email-validator/go-email-validator/pkg/ev/evmail"
	"github.com/go-email-validator/go-email-validator/pkg/ev/role"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	settingsState "github.com/osuAkatsuki/hanayo/app/states/settings"
	su "github.com/osuAkatsuki/hanayo/app/usecases/sessions"
)

// build validator without smtp check as it caused issues.
var EmailValidator = ev.NewDepBuilder(
	ev.ValidatorMap{
		ev.RoleValidatorName:       ev.NewRoleValidator(role.NewRBEASetRole()),
		ev.DisposableValidatorName: ev.NewDisposableValidator(contains.NewFunc(disposable.MailChecker)),
		ev.SyntaxValidatorName:     ev.NewSyntaxValidator(),
		ev.MXValidatorName:         ev.DefaultNewMXValidator(),
	},
).Build()

func ValidateUsername(s string) string {

	// check if violates regex
	if !usernameRegex.MatchString(s) {
		return "Your username must contain alphanumerical characters, spaces, or any of <code>_[]-</code>."
	}

	// check if username in banned names list
	lower := SafeUsername(s)
	for _, name := range forbiddenUsernames {
		if lower == SafeUsername(name) {
			return "You are not allowed to pick that username."
		}
	}

	return ""
}

func SafeUsername(u string) string {
	return strings.Replace(strings.TrimSpace(strings.ToLower(u)), " ", "_", -1)
}

func ValidateEmail(email string) bool {

	result := make(chan ev.ValidationResult)
	go func() {
		result <- EmailValidator.Validate(ev.NewInput(evmail.FromString(email)))
	}()

	return (<-result).IsValid()
}

func AddToUserNotes(message string, user int) error {
	message = "\n[" + time.Now().Format("2006-01-02") + "] " + message

	_, err := services.DB.Exec("UPDATE users SET notes = CONCAT(COALESCE(notes, ''), CONVERT(? USING utf8mb4)) WHERE id = ?",
		message, user)

	return err
}

func LogIP(c *gin.Context, user int) error {
	_, err := services.DB.Exec(
		`INSERT INTO ip_user (userid, ip, occurencies) VALUES (?, ?, '1') ON
		DUPLICATE KEY UPDATE occurencies = occurencies + 1`,
		user, su.ClientIP(c),
	)

	return err
}

// Expects structure from http://ip-api.com/json/{ip}
type Geolocation struct {
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Region      string `json:"region"`
	RegionName  string `json:"regionName"`
	City        string `json:"city"`
	Zip         string `json:"zip"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Timezone    string `json:"timezone"`
	Isp         string `json:"isp"`
	Org         string `json:"org"`
	As          string `json:"as"`
	Query       string `json:"query"`
}

func SetCountry(c *gin.Context, userID int) error {
	settings := settingsState.GetSettings()
	resp, err := http.Get(settings.IP_LOOKUP_URL + "/json/" + su.ClientIP(c))
	if err != nil {
		slog.ErrorContext(c, "error", "Could not resolve country from ip!", err.Error())
		return err
	}
	geolocation := Geolocation{}
	json.NewDecoder(resp.Body).Decode(&geolocation)
	if geolocation.CountryCode == "" || len(geolocation.CountryCode) != 2 {
		slog.ErrorContext(c, "Unknown countryCode format from ip-api.com", "code", geolocation.CountryCode)
		return nil
	}
	services.DB.Exec("UPDATE users SET country = ? WHERE id = ?", geolocation.CountryCode, userID)
	return nil
}

// This is such a mess.
var usernameRegex = regexp.MustCompile(`^[A-Za-z0-9 _\[\]-]{2,15}$`)
var forbiddenUsernames = []string{
	"rrtyui", "cookiezi", "azer", "happystick", "doomsday", "sharingan33", "andrea", "cptnxn", "reimu-desu", "hvick225", "_index",
	"my aim sucks", "kynan", "rafis", "sayonara-bye", "thelewa", "wubwoofwolf", "millhioref", "tom94", "clsw", "spectator", "exgon",
	"axarious", "angelsim", "recia", "nara", "emperorpenguin83", "bikko", "xilver", "vettel", "kuu01", "_yu68", "tasuke912",
	"dusk", "ttobas", "velperk", "jakads", "jhlee0133", "abcdullah", "yuko-", "entozer", "hdhr", "ekoro", "snowwhite", "osuplayer111",
	"musty", "nero", "elysion", "ztrot", "koreapenguin", "fort", "asphyxia", "niko", "shigetora", "kaoru", "Smoothieworld", "toy", "[toy]",
	"ozzyozrock", "fieryrage", "gosy777", "zyph", "beasttrollmc", "adamqs", "karthy", "fenrir", "rohulk", "_ryuk", "spajder", "fartownik",
	"cxu", "dunois", "ner0", "wiltchq", "-gn", "cinia pacifica", "yaong", "zeluar", "dsan", "dustice", "rucker", "firebat92", "avenging_goose",
	"idke", "vaxei", "seouless", "spare", "totoki", "rustbell", "emilia", "reimu-desu", "tiger claw", "boggles", "thepoon", "the poon", "loli_silica",
	"bahamete", "bikko", "la valse", "thelewa", "firstus", "ritzeh", "kablaze", "peppy", "loctav", "banchobot", "millhioref", "ephemeral", "flyte",
	"nanaya", "RBRat3", "smoogipoooo", "tom94", "yelle", "ztrot", "zallius", "deadbeat", "shaRPLL", "shaRPII", "shARPIL", "shARPLI", "Blaizer",
	"Damnae", "Daru", "Echo", "fly a kite", "marcin", "mm201", "nekodex", "rbrat3", "thevileone", "alumentorz", "fort", "11t", "captin1", "kroytz",
	"cryo[iceeicee]", "Akali", "professionalbox", "Fantazy", "Sing", "toybot", "goldenwolf", "handsome", "Raikozen", "cherry blossom",
	"monstrata", "Ascendence", "doorfin", "barkingmaddog", "Karen", "crystal", "vert", "halfslashed", "kloyd", "djpop", "cyclone", "guy", "sakura",
	"spectator", "pishifat", "ktgster", "skystar", "o9kami", "09kami", "Nathan", "ely", "hollow wings", "val0108", "blue dragon", "tillerino",
	"mikuia", "ameo", "tatsumaki", "cmyui", "solis", "rumoi", "frostidrinks", "cursordance", "parkes", "paparkes", "daniel", "flyingtuna",
	"walkingtuna", "nathan on osu", "justice", "child", "eb", "kalzo", "ebenezer", "solomon", "murmurtwins", "ggm9", "kaguya", "unspoken mattay",
	"mattay", "parkourwizard", "woey", "trafis", "klug", "c o i n", "varvalian", "mismagius", "nameless player", "mbmasher", "okinamo", "knalli",
	"obtio", "konnan", "ppy", "nejzha", "kochiya", "haruki", "kaguya", "miniature lamp", "phabled", "hentai", "coletaku", "zoom", "mathyu",
	"windshear", "roma4ka", "bad girl", "arfung", "skyapple", "hotzi6", "joueur de visee", "ted", "willcookie", "zerrah", "-ristuki", "yuudachi",
	"idealism", "shiiiiiii", "shayell", "parky", "torahiko", "digidrake", "a12456", "chal", "mathi", "relaxingtuna", "eriksu", "firedigger", "-hibiki-",
	"notititititi", "mysliderbreak", "qsc20010", "curry3521", "s1ck", "itswinter", "remillia", "astar", "aika", "ruri", "cpugeek", "andros",
	"xeltol", "merami", "mrekk", "whitecat", "micca", "alumetri", "fgsky", "badeu", "asecretbox", "a_blue", "lifeline", "dereban", "vamhi",
	"azr8", "azerite", "ralidea", "bartek22830", "morgn", "maxim bogdan", "gasha", "chocomint", "srchispa", "vinno", "mcy4", "arcin", "gayzmcgee",
	"filsdelama", "paraqeet", "danyl",
}
