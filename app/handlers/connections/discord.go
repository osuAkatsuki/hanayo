package connections

import (
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	eh "github.com/osuAkatsuki/hanayo/app/handlers/errors"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"

	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	"github.com/osuAkatsuki/hanayo/app/usecases/oauthstate"

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
	state, err := oauthstate.New(
		"discord",
		int(sessions.GetContext(c).User.ID),
		settings.APP_HANAYO_KEY,
		time.Now().Add(10*time.Minute),
	)
	if err != nil {
		eh.Resp500(c)
		return
	}

	discordCallbackUrl := fmt.Sprintf("%s/discord/callback", settings.PUBLIC_AKATSUKI_API_BASE_URL)
	query := url.Values{}
	query.Set("client_id", settings.DISCORD_CLIENT_ID)
	query.Set("redirect_uri", discordCallbackUrl)
	query.Set("response_type", "code")
	query.Set("scope", "identify")
	query.Set("state", state)

	c.Redirect(301, "https://discord.com/oauth2/authorize?"+query.Encode())
}

func LinkTwitchHandler(c *gin.Context) {
	if sessions.GetContext(c).User.ID == 0 {
		tu.Resp403(c)
		return
	}

	if services.DB.QueryRow("SELECT 1 FROM users WHERE id = ? AND twitch_account_id IS NOT NULL", sessions.GetContext(c).User.ID).
		Scan(new(int)) != sql.ErrNoRows {
		sessions.AddMessage(c, msg.ErrorMessage{lu.T(c, "You already have a Twitch account linked.")})
		sessions.GetSession(c).Save()

		c.Redirect(302, "/settings/connections")
		return
	}

	settings := settingsState.GetSettings()
	state, err := oauthstate.New(
		"twitch",
		int(sessions.GetContext(c).User.ID),
		settings.APP_HANAYO_KEY,
		time.Now().Add(10*time.Minute),
	)
	if err != nil {
		eh.Resp500(c)
		return
	}

	twitchCallbackUrl := fmt.Sprintf("%s/twitch/callback", settings.PUBLIC_AKATSUKI_API_BASE_URL)
	query := url.Values{}
	query.Set("client_id", settings.TWITCH_CLIENT_ID)
	query.Set("redirect_uri", twitchCallbackUrl)
	query.Set("response_type", "code")
	query.Set("state", state)

	c.Redirect(301, "https://id.twitch.tv/oauth2/authorize?"+query.Encode())
}
