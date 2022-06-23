package main

import (
	"strings"
	"strconv"

	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func mergeGET(c *gin.Context) {
	if getContext(c).User.ID == 0 {
		resp403(c)
		return
	}

	mergeResp(c)
}

func mergePOST(c *gin.Context) {
	if getContext(c).User.ID == 0 {
		resp403(c)
		return
	}

	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")
	akatsuki := strconv.Itoa(getContext(c).User.ID)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:9393/merge", nil)

	req.Header.Set("username", username)
	req.Header.Set("password", password)
	req.Header.Set("akatsuki", akatsuki)

	resp, err := client.Do(req)
	if err != nil {
		addMessage(c, errorMessage{T(c, "There was an error merging your scores. Please contact tsunyoku#0066 for assistance!")})
		c.Redirect(302, "https://akatsuki.pw")
	}

	body_b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		addMessage(c, errorMessage{T(c, "There was an error merging your scores. Please contact tsunyoku#0066 for assistance!")})
		c.Redirect(302, "https://akatsuki.pw")
	}

	body := string(body_b)

	if body == "ok" {
		mergeResp(c, successMessage{T(c, "Your scores have been queued to merge. They should be merged shortly!")})
		return
	} else if body == "password" {
		mergeResp(c, errorMessage{T(c, "Your Iteki password is wrong! Please try again.")})
		return
	} else if body == "username" {
		mergeResp(c, errorMessage{T(c, "No Iteki account was found with this username. Please try again.")})
		return
	} else if body == "no" {
		mergeResp(c, errorMessage{T(c, "You have already merged your Iteki account!")})
		return
	} else {
		mergeResp(c, errorMessage{T(c, "There was an error merging your scores. Please contact tsunyoku#0066 for assistance!")})
		return
	}
}

func mergeResp(c *gin.Context, messages ...message) {
	resp(c, 200, "merge.html", &baseTemplateData{
		TitleBar:  "Iteki Merge",
		KyutGrill: "login2.jpg",
		Messages:  messages,
		FormData:  normaliseURLValues(c.Request.PostForm),
	})
}