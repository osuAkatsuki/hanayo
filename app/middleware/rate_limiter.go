package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/sessions"
)

const reqsPerSecond = 2000
const sleepTime = time.Second / reqsPerSecond

var limiter = make(chan struct{}, reqsPerSecond)

func SetUpLimiter() {
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

func RateLimiter(onAnonymousOnly bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		if onAnonymousOnly {
			ctx := sessions.GetContext(c)
			if ctx.User.ID == 0 {
				<-limiter
			}
		} else {
			<-limiter
		}

		c.Next()
	}
}
