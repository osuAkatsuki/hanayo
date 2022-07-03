package main

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
	"fmt"

	"github.com/gin-gonic/gin"
)

func addToUserNotes(message string, user int) {
    message = "\n[" + time.Now().Format("2006-01-02") + "] " + message
	
    db.Exec("UPDATE users SET notes = CONCAT(COALESCE(notes, ''), ?) WHERE id = ?",
        message, user)
}

func changeFlag(c *gin.Context) {
	if getContext(c).User.ID == 0 {
		resp403(c)
		return
	}

	if c.PostForm("country") != "" {
		db.Exec("UPDATE users_stats SET country = ? WHERE id = ?", c.PostForm("country"), getContext(c).User.ID)
		db.Exec("UPDATE rx_stats SET country = ? WHERE id = ?", c.PostForm("country"), getContext(c).User.ID)
		rd.Publish("api:change_country", strconv.Itoa(int(getContext(c).User.ID)))
		addMessage(c, successMessage{T(c, "Flag changed")})
		getSession(c).Save()
		c.Redirect(302, "/u/"+strconv.Itoa(int(getContext(c).User.ID)))
	} else {
		addMessage(c, errorMessage{T(c, "Something went wrong.")})
		getSession(c).Save()
		c.Redirect(302, "/u/"+strconv.Itoa(int(getContext(c).User.ID)))
	}

}

func changeName(c *gin.Context) {
	if getContext(c).User.ID == 0 {
		resp403(c)
		return
	}
	if c.PostForm("name") != "" {
		username := strings.TrimSpace(c.PostForm("name"))
		// check if username already taken
		if db.QueryRow("SELECT 1 FROM users WHERE username_safe = ?", safeUsername(username)).
			Scan(new(int)) != sql.ErrNoRows {
			addMessage(c, errorMessage{T(c, "Username taken.")})
			getSession(c).Save()
			c.Redirect(302, "/u/"+strconv.Itoa(int(getContext(c).User.ID)))
			return
		}
		// check if violates regex
		if !usernameRegex.MatchString(username) {
			addMessage(c, errorMessage{T(c, "Please choose a username that matches our criteria.")})
			getSession(c).Save()
			c.Redirect(302, "/u/"+strconv.Itoa(int(getContext(c).User.ID)))

			return
		}
		// check if username in banned names list c
		if in(strings.ToLower(username), forbiddenUsernames) {
			addMessage(c, errorMessage{T(c, "You are not allowed to pick that username.")})
			getSession(c).Save()
			c.Redirect(302, "/u/"+strconv.Itoa(int(getContext(c).User.ID)))

			return
		}

		// update username
		db.Exec("UPDATE users_stats SET username = ? WHERE id = ?", username, getContext(c).User.ID)
		db.Exec("UPDATE rx_stats SET username = ? WHERE id = ?", username, getContext(c).User.ID)
		db.Exec("UPDATE users SET username = ?, username_safe = ? WHERE id = ?", username, safeUsername(username), getContext(c).User.ID)
		addToUserNotes(fmt.Sprintf("Username change: %s -> %s", getContext(c).User.Username, username), getContext(c).User.ID)
		rd.Publish("api:change_username", strconv.Itoa(int(getContext(c).User.ID)))
		addMessage(c, successMessage{T(c, "Username changed")})
		getSession(c).Save()
		c.Redirect(302, "/u/"+strconv.Itoa(int(getContext(c).User.ID)))
	} else {
		addMessage(c, errorMessage{T(c, "Something went wrong.")})
		getSession(c).Save()
		c.Redirect(302, "/u/"+strconv.Itoa(int(getContext(c).User.ID)))
	}
}
