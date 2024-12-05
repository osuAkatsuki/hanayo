package connections

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"

	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"

	msg "github.com/osuAkatsuki/hanayo/app/models/messages"

	"github.com/osuAkatsuki/hanayo/app/states/services"
	settingsState "github.com/osuAkatsuki/hanayo/app/states/settings"
)

func LinkDiscordHandler(c *gin.Context) {
	if sessions.GetContext(c).User.ID == 0 {
		tu.Resp403(c)
		return
	}

	if services.DB.QueryRow("SELECT 1 FROM users WHERE id = ? AND discord_account_id IS NOT NULL", sessions.GetContext(c).User.ID).
		Scan(new(int)) != sql.ErrNoRows {
		sessions.AddMessage(c, msg.ErrorMessage{lu.T(c, "You already have a Discord account linked.")})
		sessions.GetSession(c).Save()

		c.Redirect(302, "/u/"+strconv.Itoa(int(sessions.GetContext(c).User.ID)))
		return
	}

	settings := settingsState.GetSettings()

	discordCallbackUrl := fmt.Sprintf("%s/discord/callback", settings.PUBLIC_AKATSUKI_API_BASE_URL)
	redirectUrl := fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=identify", settings.DISCORD_CLIENT_ID, discordCallbackUrl)

	c.Redirect(301, redirectUrl)
}
