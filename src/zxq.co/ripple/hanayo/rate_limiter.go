package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

const reqsPerSecond = 2000
const sleepTime = time.Second / reqsPerSecond

var limiter = make(chan struct{}, reqsPerSecond)

func setUpLimiter() {
	for i := 0; i < 2000; i++ {
		limiter <- struct{}{}
	}
	go func() {
		for {
			limiter <- struct{}{}
			time.Sleep(sleepTime)
		}
	}()
}

func rateLimiter(onAnonymousOnly bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		if onAnonymousOnly {
			ctx := getContext(c)
			if ctx.User.ID == 0 {
				<-limiter
			}
		} else {
			<-limiter
		}

		c.Next()
	}
}
