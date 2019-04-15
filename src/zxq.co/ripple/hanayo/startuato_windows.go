// +build windows

package main

import (
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

var l net.Listener

func startuato(engine *gin.Engine) {
	var err error

	// Listen on a TCP or a UNIX domain socket (TCP here).
	if config.Unix {
		l, err = net.Listen("unix", config.ListenTo)
	} else {
		l, err = net.Listen("tcp", config.ListenTo)
	}
	if err != nil {
		log.Fatalln(err)
	}

	http.Serve(l, engine)
}
