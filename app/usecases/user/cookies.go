package user

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/akatsuki-api/common"
	"github.com/osuAkatsuki/hanayo/app/states/services"
)

func SetYCookie(userID int, c *gin.Context) {
	var token string
	err := services.DB.QueryRow("SELECT token FROM identity_tokens WHERE userid = ? LIMIT 1", userID).Scan(&token)
	if err != nil && err != sql.ErrNoRows {
		c.Error(err)
		return
	}
	if token != "" {
		AddY(c, token)
		return
	}
	for {
		token = fmt.Sprintf("%x", sha256.Sum256([]byte(common.RandomString(32))))
		if services.DB.QueryRow("SELECT 1 FROM identity_tokens WHERE token = ? LIMIT 1", token).Scan(new(int)) == sql.ErrNoRows {
			break
		}
	}
	services.DB.Exec("INSERT INTO identity_tokens(userid, token) VALUES (?, ?)", userID, token)
	AddY(c, token)
}

func AddY(c *gin.Context, y string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:    "y",
		Value:   y,
		Expires: time.Now().Add(time.Hour * 24 * 30 * 6),
	})
}
