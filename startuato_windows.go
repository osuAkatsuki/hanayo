//go:build windows
// +build windows

package main

import (
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/osuAkatsuki/hanayo/app/states/listener"
	"github.com/osuAkatsuki/hanayo/app/states/settings"
)

func startuato(engine *gin.Engine) bool {
	var err error

	// Listen on a TCP or a UNIX domain socket (TCP here).
	if config.Unix {
		listener.Listener, err = net.Listen("unix", settings.Config.ListenTo)
	} else {
		listener.Listener, err = net.Listen("tcp", settings.Config.ListenTo)
	}
	if err != nil {
		log.Fatalln(err)
		return false
	}

	http.Serve(listener.Listener, engine)
	return true
}
