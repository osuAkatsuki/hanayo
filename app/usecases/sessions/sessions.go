package sessions

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/akatsuki-api/common"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	"github.com/osuAkatsuki/hanayo/app/usecases/auth/cryptography"
)

func ClientIP(c *gin.Context) string {
	ff := c.Request.Header.Get("CF-Connecting-IP")
	if ff != "" {
		return ff
	}
	return c.ClientIP()
}

func GenerateToken(id int, c *gin.Context) (string, error) {
	tok := common.RandomString(32)
	_, err := services.DB.Exec(
		`INSERT INTO tokens(user, privileges, description, token, private)
					VALUES (   ?,        '0',           ?,     ?,     '1');`,
		id, ClientIP(c), cryptography.MakeMD5(tok))
	if err != nil {
		return "", err
	}
	return tok, nil
}

func CheckToken(s string, id int, c *gin.Context) (string, error) {
	if s == "" {
		return GenerateToken(id, c)
	}

	if err := services.DB.QueryRow("SELECT 1 FROM tokens WHERE token = ?", cryptography.MakeMD5(s)).Scan(new(int)); err == sql.ErrNoRows {
		return GenerateToken(id, c)
	} else if err != nil {
		return "", err
	}

	return s, nil
}
