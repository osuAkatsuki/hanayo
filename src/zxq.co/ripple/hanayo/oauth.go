package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"zxq.co/ripple/hanayo/routers/oauth"
)

type oauthRequestHandler struct{}

func (o oauthRequestHandler) CheckLoggedInOrRedirect(c *gin.Context) bool {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		resp403(c)
		return false
	}
	return true
}

func (o oauthRequestHandler) DisplayAuthorizePage(client oauth.Client, c *gin.Context) {
	var creatorName string
	db.Get(&creatorName, "SELECT username FROM users WHERE id = ? LIMIT 1", client.CreatorID)

	c.Header("X-Frame-Options", "deny")

	simple(c, getSimpleByFilename("oauth.html"), nil, map[string]interface{}{
		"Client":      client,
		"CreatorName": creatorName,
	})
}

func (o oauthRequestHandler) CheckCSRF(c *gin.Context, s string) bool {
	b, _ := CSRF.Validate(getContext(c).User.ID, s)
	return b
}

func (o oauthRequestHandler) GetDB() *sql.DB {
	return db.DB
}

func (o oauthRequestHandler) GetUserID(c *gin.Context) int {
	return getContext(c).User.ID
}

func setUpOauth() {
	oauth.Initialise(oauthRequestHandler{})
}
