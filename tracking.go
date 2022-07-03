package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"zxq.co/x/rs"
	"github.com/gin-gonic/gin"
)

func setYCookie(userID int, c *gin.Context) {
	var token string
	err := db.QueryRow("SELECT token FROM identity_tokens WHERE userid = ? LIMIT 1", userID).Scan(&token)
	if err != nil && err != sql.ErrNoRows {
		c.Error(err)
		return
	}
	if token != "" {
		addY(c, token)
		return
	}
	for {
		token = fmt.Sprintf("%x", sha256.Sum256([]byte(rs.String(32))))
		if db.QueryRow("SELECT 1 FROM identity_tokens WHERE token = ? LIMIT 1", token).Scan(new(int)) == sql.ErrNoRows {
			break
		}
	}
	db.Exec("INSERT INTO identity_tokens(userid, token) VALUES (?, ?)", userID, token)
	addY(c, token)
}
func addY(c *gin.Context, y string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:    "y",
		Value:   y,
		Expires: time.Now().Add(time.Hour * 24 * 30 * 6),
	})
}

func logIP(c *gin.Context, user int) {
	db.Exec(`INSERT INTO ip_user (userid, ip, occurencies) VALUES (?, ?, '1')
						ON DUPLICATE KEY UPDATE occurencies = occurencies + 1`, user, clientIP(c))
}

func setCountry(c *gin.Context, user int) error {
	raw, err := http.Get(config.IP_API + "/" + clientIP(c) + "/country")
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(raw.Body)
	if err != nil {
		return err
	}
	country := strings.TrimSpace(string(data))
	if country == "" || len(country) != 2 {
		return nil
	}
	db.Exec("UPDATE users_stats SET country = ? WHERE id = ?", country, user)
	return nil
}
