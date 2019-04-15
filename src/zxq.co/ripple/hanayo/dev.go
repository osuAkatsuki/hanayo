package main

import (
	"crypto/md5"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	"zxq.co/ripple/rippleapi/common"
	"zxq.co/x/rs"
)

func createAPIToken(c *gin.Context) {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		resp403(c)
		return
	}

	sess := getSession(c)
	defer func() {
		sess.Save()
		c.Redirect(302, "/dev/tokens")
	}()

	if ok, _ := CSRF.Validate(ctx.User.ID, c.PostForm("csrf")); !ok {
		addMessage(c, errorMessage{T(c, "Your session has expired. Please try redoing what you were trying to do.")})
		return
	}

	privileges := common.Privileges(common.Int(c.PostForm("privileges"))).CanOnly(ctx.User.Privileges)
	description := c.PostForm("description")

	var (
		tokenStr string
		tokenMD5 string
	)

	for {
		tokenStr = common.RandomString(32)
		tokenMD5 = fmt.Sprintf("%x", md5.Sum([]byte(tokenStr)))

		var id int
		err := db.QueryRow("SELECT id FROM tokens WHERE token = ? LIMIT 1", tokenMD5).Scan(&id)
		if err == sql.ErrNoRows {
			break
		}
		if err != nil {
			c.Error(err)
			resp500(c)
			return
		}
	}

	_, err := db.Exec("INSERT INTO tokens(user, privileges, description, token, private, last_updated) VALUES (?, ?, ?, ?, '0', ?)",
		ctx.User.ID, privileges, description, tokenMD5, time.Now().Unix())
	if err != nil {
		c.Error(err)
		resp500(c)
		return
	}

	addMessage(c, successMessage{
		fmt.Sprintf("Your token has been created successfully! Your token is: <code>%s</code>.<br>Keep it safe, don't show it around, and store it now! We won't show it to you again.", tokenStr),
	})
}

func deleteAPIToken(c *gin.Context) {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		resp403(c)
		return
	}

	sess := getSession(c)
	defer func() {
		sess.Save()
		c.Redirect(302, "/dev/tokens")
	}()

	if ok, _ := CSRF.Validate(ctx.User.ID, c.PostForm("csrf")); !ok {
		addMessage(c, errorMessage{T(c, "Your session has expired. Please try redoing what you were trying to do.")})
		return
	}

	db.Exec("DELETE FROM tokens WHERE id = ? AND user = ? AND private = 0", c.PostForm("id"), ctx.User.ID)
	addMessage(c, successMessage{"That token has been deleted successfully."})
}

func editAPIToken(c *gin.Context) {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		resp403(c)
		return
	}

	sess := getSession(c)
	defer func() {
		sess.Save()
		c.Redirect(302, "/dev/tokens/edit?id="+c.PostForm("id"))
	}()

	if ok, _ := CSRF.Validate(ctx.User.ID, c.PostForm("csrf")); !ok {
		addMessage(c, errorMessage{T(c, "Your session has expired. Please try redoing what you were trying to do.")})
		return
	}

	privileges := common.Privileges(common.Int(c.PostForm("privileges"))).CanOnly(ctx.User.Privileges)
	description := c.PostForm("description")

	_, err := db.Exec("UPDATE tokens SET privileges = ?, description = ? WHERE user = ? AND id = ? AND private = 0",
		privileges, description, ctx.User.ID, c.PostForm("id"))
	if err != nil {
		c.Error(err)
		resp500(c)
		return
	}

	addMessage(c, successMessage{
		"Your token has been edited successfully!",
	})
}

type oAuthClient struct {
	ID          string
	Extra       string
	RedirectURI string
}

// Name retrieves the name of an oAuthClient
func (o oAuthClient) Name() string {
	var x [3]string
	json.Unmarshal([]byte(o.Extra), &x)
	return x[0]
}

// Avatar retrieves the avatar of an oAuthClient
func (o oAuthClient) Avatar() string {
	var x [3]string
	json.Unmarshal([]byte(o.Extra), &x)
	return x[2]
}

// Owner retrieves the ID of the owner of an oAuthClient
func (o oAuthClient) Owner() int {
	var x [3]string
	json.Unmarshal([]byte(o.Extra), &x)
	u, _ := strconv.Atoi(x[1])
	return u
}

func getOAuthApplications(c *gin.Context) {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		resp403(c)
		return
	}

	var apps []oAuthClient
	err := db.Select(&apps, `SELECT c.id, c.extra, c.redirect_uri as redirecturi
FROM osin_client_user cu
INNER JOIN osin_client c ON c.id = cu.client_id
WHERE cu.user = ?`, ctx.User.ID)

	if err != nil {
		c.Error(err)
		resp500(c)
		return
	}

	simple(c, getSimpleByFilename("dev/apps.html"), nil, map[string]interface{}{
		"apps": apps,
	})
}

func editOAuthApplication(c *gin.Context) {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		resp403(c)
		return
	}

	app := new(oAuthClient)
	if c.Query("id") != "new" {
		err := db.Get(app, `SELECT c.id, c.extra, c.redirect_uri as redirecturi
FROM osin_client_user cu
INNER JOIN osin_client c ON c.id = cu.client_id
WHERE cu.user = ? AND cu.client_id = ?`, ctx.User.ID, c.Query("id"))
		switch err {
		case nil:
			break
		case sql.ErrNoRows:
			app = nil
		default:
			c.Error(err)
			resp500(c)
			return
		}
	} else {
		app = new(oAuthClient)
	}

	tpl := getSimpleByFilename("dev/edit_app.html")
	if c.Query("id") == "new" {
		tpl.TitleBar = "Create OAuth 2 application"
	}
	simple(c, tpl, nil, map[string]interface{}{
		"app": app,
	})
}

func editOAuthApplicationSubmit(c *gin.Context) {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		resp403(c)
		return
	}

	c.Request.ParseMultipartForm(32 << 10)
	id := c.PostForm("id")

	sess := getSession(c)
	defer func() {
		sess.Save()
		c.Redirect(302, "/dev/apps/edit?id="+id)
	}()

	var avatarFilename string
	if id != "new" {
		var previousExtra string
		db.Get(&previousExtra, "SELECT extra FROM osin_client WHERE id = ?", id)
		oClient := oAuthClient{Extra: previousExtra}
		if previousExtra == "" || oClient.Owner() != ctx.User.ID {
			fmt.Println(previousExtra, oClient.Owner(), ctx.User.ID)
			// user will be returned to /dev/apps/edit?id=whatever,
			// which will say the token could not be found.
			return
		}
		avatarFilename = oClient.Avatar()
	}

	avatar, _, err := c.Request.FormFile("avatar")

	if err != nil && err != http.ErrMissingFile {
		c.Error(err)
		resp500(c)
		return
	}

	if avatar != nil {
		if avatarFilename != "" {
			err := deletePreviousAvatar(avatarFilename)
			if err != nil {
				c.Error(err)
				resp500(c)
				return
			}
		}
		avatarFilename, err = storeNewAvatar(avatar)
		if err != nil {
			c.Error(err)
			resp500(c)
			return
		}
	}

	name := common.SanitiseString(c.PostForm("name"))
	if len(name) > 25 {
		name = name[:25]
	}
	redirectURI := common.SanitiseString(c.PostForm("redirect_uri"))
	extra, _ := json.Marshal([3]string{
		name,
		strconv.Itoa(ctx.User.ID),
		avatarFilename,
	})

	if id == "new" {
		id = rs.StringFromChars(32, oAuthRandomChars)
		secret := rs.StringFromChars(64, oAuthRandomChars)
		secretSha := fmt.Sprintf("%x", sha256.Sum256([]byte(secret)))
		db.Exec("INSERT INTO osin_client(id, secret, extra, redirect_uri) VALUES (?, ?, ?, ?)", id, secretSha, string(extra), redirectURI)
		db.Exec("INSERT INTO osin_client_user(client_id, user) VALUES (?, ?)", id, ctx.User.ID)
		addMessage(c, successMessage{fmt.Sprintf(createClientMessage, id, secret)})
	} else {
		db.Exec("UPDATE osin_client SET extra = ?, redirect_uri = ? WHERE id = ?", string(extra), redirectURI, id)
		addMessage(c, successMessage{"Your application has been saved."})
	}
}

const createClientMessage = `
You can now get going integrating Ripple in your super cool project!
Here's what you need:<br>
<pre>client_id     = "%s"
client_secret = "%s"</pre>
As always: keep it safe, don't show it around, and store it now!
We won't show you the client_secret again.`

func deletePreviousAvatar(s string) error {
	return os.Remove("static/oauth-apps/" + s)
}

const oAuthRandomChars = "qwertyuiopasdfghjklzxcvbnm"

func storeNewAvatar(file multipart.File) (string, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}
	img = resize.Thumbnail(256, 256, img, resize.Bilinear)
	newFilename := rs.StringFromChars(32, oAuthRandomChars) + ".png"
	f, err := os.Create(fmt.Sprintf("static/oauth-apps/%s", newFilename))
	defer f.Close()
	if err != nil {
		return "", err
	}
	err = png.Encode(f, img)
	return newFilename, err
}

func deleteOAuthApplication(c *gin.Context) {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		resp403(c)
		return
	}

	sess := getSession(c)
	defer func() {
		sess.Save()
		c.Redirect(302, "/dev/apps")
	}()

	if ok, _ := CSRF.Validate(ctx.User.ID, c.PostForm("csrf")); !ok {
		addMessage(c, errorMessage{T(c, "Your session has expired. Please try redoing what you were trying to do.")})
		return
	}

	var x oAuthClient
	db.Get(&x, "SELECT extra FROM osin_client WHERE id = ?", c.PostForm("id"))
	if x.Owner() != ctx.User.ID {
		addMessage(c, errorMessage{"y u do dis"})
		return
	}

	if x.Avatar() != "" {
		os.Remove("static/oauth-apps/" + x.Avatar())
	}

	clientID := c.PostForm("id")

	db.Exec("DELETE FROM osin_access WHERE client = ?", clientID)
	db.Exec("DELETE FROM osin_client_user WHERE client_id = ?", clientID)
	db.Exec("DELETE FROM osin_client WHERE id = ?", clientID)

	addMessage(c, successMessage{"poof"})
}

type authorization struct {
	oAuthClient
	Scope     string
	CreatedAt time.Time
	Client    string
}

var scopeMap = map[string]string{
	"identify":          "Identify",
	"read_confidential": "Read private information",
	"write":             "Write",
}

func (a authorization) Scopes(c *gin.Context) string {
	if a.Scope == "" {
		return T(c, "Identify")
	}
	scopes := strings.Split(a.Scope, " ")
	scopes = append([]string{"identify"}, scopes...)
	for i, val := range scopes {
		scopes[i] = T(c, scopeMap[val])
	}
	return strings.Join(scopes, ", ")
}

func authorizedApplications(c *gin.Context) {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		resp403(c)
	}

	var apps []authorization
	err := db.Select(&apps, `
SELECT c.extra, a.scope, a.created_at AS createdat, a.client
FROM osin_access a
INNER JOIN osin_client c ON c.id = a.client
WHERE a.extra = ?
GROUP BY a.client
ORDER BY a.created_at DESC`, ctx.User.ID)

	if err != nil {
		c.Error(err)
		resp500(c)
		return
	}

	simple(
		c, getSimpleByFilename("settings/authorized_applications.html"), nil,
		map[string]interface{}{
			"apps": apps,
		},
	)
}

func revokeAuthorization(c *gin.Context) {
	ctx := getContext(c)
	if ctx.User.ID == 0 {
		resp403(c)
		return
	}

	sess := getSession(c)
	defer func() {
		sess.Save()
		c.Redirect(302, "/settings/authorized_applications")
	}()

	if ok, _ := CSRF.Validate(ctx.User.ID, c.PostForm("csrf")); !ok {
		addMessage(c, errorMessage{T(c, "Your session has expired. Please try redoing what you were trying to do.")})
		return
	}

	db.Exec("DELETE FROM osin_access WHERE client = ? AND extra = ?", c.PostForm("client_id"), ctx.User.ID)
	addMessage(c, successMessage{T(c, "That authorization has been successfully revoked.")})
}
