package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	raven "github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
)

// Recovery is a better sentry logger.
func Recovery(client *raven.Client, onlyCrashes bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody []byte

		defer func() {
			st := raven.NewStacktrace(0, 3, []string{"git.zxq.co/ripple"})

			tokenRaw, ex := c.Get("token")
			var token string
			if ex {
				token = tokenRaw.(string)
			}

			ravenHTTP := raven.NewHttp(c.Request)
			if len(requestBody) != 0 {
				ravenHTTP.Data = string(requestBody)
			}

			ravenUser := &raven.User{
				Username: token,
				IP:       c.Request.RemoteAddr,
			}

			flags := map[string]string{
				"endpoint": c.Request.RequestURI,
				"token":    token,
			}

			if rval := recover(); rval != nil {
				var err error
				switch rval := rval.(type) {
				case string:
					err = errors.New(rval)
				case error:
					err = rval
				default:
					err = fmt.Errorf("%v - %#v", rval, rval)
				}
				fmt.Println(err)
				client.CaptureError(
					err,
					flags,
					st,
					ravenHTTP,
					ravenUser,
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			if !onlyCrashes {
				for _, item := range c.Errors {
					var err = error(item)
					if item.Type == gin.ErrorTypePrivate {
						err = item.Err
					}
					fmt.Println(err)
					client.CaptureError(
						err,
						flags,
						ravenHTTP,
						ravenUser,
					)
				}
			}
		}()

		if c.Request.Method == "POST" && c.Request.URL.Path != "/tokens" &&
			c.Request.URL.Path != "/tokens/new" {
			var err error
			requestBody, err = ioutil.ReadAll(c.Request.Body)
			if err != nil {
				c.Error(err)
			}
			c.Request.Body = fakeBody{
				r:    bytes.NewReader(requestBody),
				orig: c.Request.Body,
			}
		}

		c.Next()
	}
}

type fakeBody struct {
	r    io.Reader
	orig io.ReadCloser
}

func (f fakeBody) Read(p []byte) (int, error) { return f.r.Read(p) }
func (f fakeBody) Close() error               { return f.orig.Close() }
