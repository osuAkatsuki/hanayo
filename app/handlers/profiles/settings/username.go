package settings

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/slog"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/akatsuki-api/common"
	msg "github.com/osuAkatsuki/hanayo/app/models/messages"
	"github.com/osuAkatsuki/hanayo/app/sessions"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	lu "github.com/osuAkatsuki/hanayo/app/usecases/localisation"
	tu "github.com/osuAkatsuki/hanayo/app/usecases/templates"
	uu "github.com/osuAkatsuki/hanayo/app/usecases/user"
)

func NameChangeSubmitHandler(c *gin.Context) {
	ctx := sessions.GetContext(c)
	if ctx.User.ID == 0 {
		tu.Resp403(c)
		return
	}

	if ctx.User.Privileges&common.UserPrivilegeDonor == 0 {
		tu.Resp403(c)
		return
	}

	if c.PostForm("name") == "" {
		sessions.AddMessage(c, msg.ErrorMessage{lu.T(c, "Something went wrong.")})
		sessions.GetSession(c).Save()
		c.Redirect(302, "/u/"+strconv.Itoa(int(ctx.User.ID)))
		return
	}

	username := strings.TrimSpace(c.PostForm("name"))
	err := uu.ValidateUsername(username)
	if err != "" {
		sessions.AddMessage(c, msg.ErrorMessage{lu.T(c, err)})
		sessions.GetSession(c).Save()
		c.Redirect(302, "/u/"+strconv.Itoa(int(ctx.User.ID)))
		return
	}
	
	safeUsername := uu.SafeUsername(username)

	// check if username already taken
	if services.DB.QueryRow("SELECT 1 FROM users WHERE username_safe = ?", safeUsername).Scan(new(int)) != sql.ErrNoRows {
		sessions.AddMessage(c, msg.ErrorMessage{lu.T(c, "Username taken.")})
		sessions.GetSession(c).Save()
		c.Redirect(302, "/u/"+strconv.Itoa(int(ctx.User.ID)))
		return
	}

	// update username
	services.DB.Exec("UPDATE users SET username = ?, username_safe = ? WHERE id = ?", username, safeUsername, ctx.User.ID)

	logErr := uu.AddToUserNotes(fmt.Sprintf("Username change (self): %s -> %s", sessions.GetContext(c).User.Username, username), sessions.GetContext(c).User.ID)
	if logErr != nil {
		slog.Error("Error adding to user notes", "error", logErr.Error())
		sessions.AddMessage(c, msg.ErrorMessage{lu.T(c, "Something went wrong.")})
		sessions.GetSession(c).Save()
		c.Redirect(302, "/u/"+strconv.Itoa(int(sessions.GetContext(c).User.ID)))
		return
	}

	services.RD.Publish("api:change_username", strconv.Itoa(int(sessions.GetContext(c).User.ID)))
	
	sessions.AddMessage(c, msg.SuccessMessage{lu.T(c, "Username changed")})
	sessions.GetSession(c).Save()
	c.Redirect(302, "/u/"+strconv.Itoa(int(sessions.GetContext(c).User.ID)))
}
