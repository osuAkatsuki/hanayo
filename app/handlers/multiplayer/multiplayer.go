package multiplayer

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"

	"golang.org/x/exp/slog"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/akatsuki-api/common"
	"github.com/osuAkatsuki/hanayo/app/models"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
)

type JsonList []int

func (jl *JsonList) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &jl)
		return nil
	case string:
		json.Unmarshal([]byte(v), &jl)
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}
func (jl *JsonList) Value() (driver.Value, error) {
	return json.Marshal(jl)
}

func inList(list []int, value int) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func MultiplayerHistoryHandler(c *gin.Context) {
	var (
		participantsIds JsonList
		privateMatch    bool
	)
	data := new(models.MatchData)
	ctx := sessions.GetContext(c)
	defer tu.Resp(c, 200, "multiplayer.html", data)

	mid := c.Param("mid")
	if _, err := strconv.Atoi(mid); err != nil {
		c.Error(err)
		slog.ErrorContext(c, err.Error())
	} else {
		err := services.DB.QueryRow(
			`SELECT id, name, private FROM matches WHERE id = ? LIMIT 1`, mid,
		).Scan(&data.MatchID, &data.MatchName, &privateMatch)

		if err != nil {
			c.Error(err)
			slog.ErrorContext(c, err.Error())
		}

		err = services.DB.Get(
			&participantsIds,
			`SELECT JSON_ARRAYAGG(user_id) FROM (
				SELECT DISTINCT(user_id) AS user_id FROM 
				match_events WHERE match_id = ? AND event_type = 'MATCH_USER_JOIN'
			) AS events`,
			data.MatchID,
		)

		if err != nil {
			c.Error(err)
			slog.ErrorContext(c, err.Error())
		}
	}

	data.BannerContent = "/static/images/headers/2fa.jpg"
	data.BannerType = 1

	if data.MatchID == 0 {
		data.TitleBar = lu.T(c, "Match not found.")
		data.Messages = append(data.Messages, msg.ErrorMessage{lu.T(c, "Match could not be found.")})
		return
	}

	if privateMatch &&
		(!inList(participantsIds, ctx.User.ID) ||
			ctx.User.Privileges&common.UserPrivilegeTournamentStaff == 0) {
		data.MatchID = 0
		data.TitleBar = lu.T(c, "Match not found.")
		data.Messages = append(data.Messages, msg.ErrorMessage{lu.T(c, "Match could not be found.")})
		return
	}

	data.TitleBar = lu.T(c, "Matches")
	data.Scripts = append(data.Scripts, "/static/js/pages/multiplayer.js")
}
