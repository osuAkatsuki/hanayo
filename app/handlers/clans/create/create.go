package create

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/akatsuki-api/common"
	"github.com/osuAkatsuki/hanayo/app/models"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	"github.com/osuAkatsuki/hanayo/app/usecases/misc"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

func ClanCreatePageHandler(c *gin.Context) {
	clanCreateResp(c)
}

func ClanCreateSubmitHandler(c *gin.Context) {
	if sessions.GetContext(c).User.ID == 0 {
		tu.Resp403(c)
		return
	}
	// check registrations are enabled
	if !clanCreationEnabled() {
		clanCreateResp(c, msg.ErrorMessage{lu.T(c, "Sorry, it's not possible to create a clan at the moment. Please try again later.")})
		return
	}

	// check username is valid by our criteria
	username := strings.TrimSpace(c.PostForm("username"))
	if !clanNameRegex.MatchString(username) {
		clanCreateResp(c, msg.ErrorMessage{lu.T(c, "Your clans name must contain alphanumerical characters, spaces, or any of <code>_[]-</code>.")})
		return
	}

	// check whether name already exists
	if services.DB.QueryRow("SELECT 1 FROM clans WHERE name = ?", c.PostForm("username")).
		Scan(new(int)) != sql.ErrNoRows {
		clanCreateResp(c, msg.ErrorMessage{lu.T(c, "A clan with that name already exists!")})
		return
	}

	// check whether tag already exists
	if services.DB.QueryRow("SELECT 1 FROM clans WHERE tag = ?", c.PostForm("tag")).
		Scan(new(int)) != sql.ErrNoRows {
		clanCreateResp(c, msg.ErrorMessage{lu.T(c, "A clan with that tag already exists!")})
		return
	}

	// recaptcha verify

	tag := "0"
	if c.PostForm("tag") != "" {
		tag = c.PostForm("tag")
	}

	// The actual registration.

	invite := common.RandomString(8)

	for services.DB.QueryRow("SELECT 1 FROM clans WHERE invite = ?", invite).Scan(new(int)) != sql.ErrNoRows {
		invite = common.RandomString(8)
	}

	res, err := services.DB.Exec(`INSERT INTO clans(name, description, icon, tag, owner, invite, status)
							  VALUES (?, ?, ?, ?, ?, ?, 2);`,
		username, c.PostForm("password"), c.PostForm("email"), tag, sessions.GetContext(c).User.ID, invite)
	if err != nil {
		clanCreateResp(c, msg.ErrorMessage{lu.T(c, "Whoops, an error slipped in. Clan might have been created, though. I don't know.")})
		fmt.Println(err)
		return
	}
	lid, _ := res.LastInsertId()

	services.DB.Exec("UPDATE users SET clan_id = ?, clan_privileges = 8 WHERE id = ?", lid, sessions.GetContext(c).User.ID)

	sessions.AddMessage(c, msg.SuccessMessage{lu.T(c, "Clan created.")})
	sessions.GetSession(c).Save()
	c.Redirect(302, "/c/"+strconv.Itoa(int(lid)))
}

func clanCreateResp(c *gin.Context, messages ...msg.Message) {
	tu.Resp(c, 200, "clans/create.html", &models.BaseTemplateData{
		TitleBar:      "Create Clan",
		BannerContent: "register.jpg",
		BannerType:    1,
		Messages:      messages,
		FormData:      misc.NormaliseURLValues(c.Request.PostForm),
	})
}

func clanCreationEnabled() bool {
	var enabled bool
	services.DB.QueryRow("SELECT value_int FROM system_settings WHERE name = 'ccreation_enabled'").Scan(&enabled)
	return enabled
}

var clanNameRegex = regexp.MustCompile(`^[A-Za-z0-9 '_\[\]-]{2,15}$`)
