//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/states/listener"
	"github.com/osuAkatsuki/hanayo/app/states/services"
	"github.com/osuAkatsuki/hanayo/app/states/settings"
	"github.com/rcrowley/goagain"
	schiavo "zxq.co/ripple/schiavolib"
)

func startuato(engine *gin.Engine) bool {

	returnCh := make(chan bool)
	// whether it was from this very thing or not
	var iZingri bool
	hs := func(l net.Listener, h http.Handler) {
		err := http.Serve(l, h)
		if f, ok := err.(*net.OpError); ok && f.Err.Error() == "use of closed network connection" && !iZingri {
			returnCh <- true
		}
	}

	var err error
	// Inherit a net.Listener from our parent process or listen anew.
	listener.Listener, err = goagain.Listener()
	if err != nil {

		// Listen on a TCP or a UNIX domain socket (TCP here).
		if settings.Config.Unix {
			listener.Listener, err = net.Listen("unix", settings.Config.ListenTo)
		} else {
			listener.Listener, err = net.Listen("tcp", settings.Config.ListenTo)
		}
		if err != nil {
			schiavo.Bunker.Send(err.Error())
			log.Fatalln(err)
		}

		schiavo.Bunker.Send(fmt.Sprint("LISTENINGU STARTUATO ON ", listener.Listener.Addr()))

		// Accept connections in a new goroutine.
		go hs(listener.Listener, engine)

	} else {

		// Resume accepting connections in a new goroutine.
		schiavo.Bunker.Send(fmt.Sprint("LISTENINGU RESUMINGU ON ", listener.Listener.Addr()))
		go hs(listener.Listener, engine)

		// Kill the parent, now that the child has started successfully.
		if err := goagain.Kill(); err != nil {
			schiavo.Bunker.Send(err.Error())
			log.Fatalln(err)
		}

	}

	go func() {
		// Block the main goroutine awaiting signals.
		if _, err := goagain.Wait(listener.Listener); err != nil {
			schiavo.Bunker.Send(err.Error())
			log.Fatalln(err)
		}

		// Do whatever's necessary to ensure a graceful exit like waiting for
		// goroutines to terminate or a channel to become closed.
		//
		// In this case, we'll simply stop listening and wait one second.
		iZingri = true
		if err := listener.Listener.Close(); err != nil {
			schiavo.Bunker.Send(err.Error())
			log.Fatalln(err)
		}
		if err := services.DB.Close(); err != nil {
			schiavo.Bunker.Send(err.Error())
			log.Fatalln(err)
		}
		returnCh <- false
	}()

	return <-returnCh
}
