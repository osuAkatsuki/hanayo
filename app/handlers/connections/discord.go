package connections

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log/slog"
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

	state, err := createStateString()

	if err != nil {
		sessions.AddMessage(c, msg.ErrorMessage{lu.T(c, "Something went wrong.")})
		sessions.GetSession(c).Save()

		slog.Error("Error generating string for Discord connection", "error", err.Error())

		c.Redirect(302, "/settings/connections")
		return
	}

	_, err = services.DB.Exec("INSERT INTO discord_states (user_id, state) VALUES (?, ?)", sessions.GetContext(c).User.ID, state)

	if err != nil {
		sessions.AddMessage(c, msg.ErrorMessage{lu.T(c, "Something went wrong.")})
		sessions.GetSession(c).Save()

		slog.Error("Error inserting to discord_states table", "error", err.Error())

		c.Redirect(302, "/settings/connections")
		return
	}

	settings := settingsState.GetSettings()

	discordCallbackUrl := fmt.Sprintf("%s/discord/callback", settings.PUBLIC_AKATSUKI_API_BASE_URL)
	rediectUrl := fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=identify&state=%s", settings.DISCORD_CLIENT_ID, discordCallbackUrl, state)

	c.Redirect(301, rediectUrl)
}

func createStateString() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}
