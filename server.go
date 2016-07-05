package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ellcrys/openmint/www"
	"github.com/ellcrys/util"
	"github.com/labstack/echo/engine/standard"
)

func main() {

	// determine appropriate port number
	var portEnv = util.Env("PORT", "3001")
	var portFlag = flag.String("port", portEnv, "set port. Default: "+portEnv)
	flag.Parse()

	if flag.Parsed() {

		// create new router
		router, _ := www.App(false, false)
		server := standard.New(":" + *portFlag)
		server.SetHandler(router)

		time.AfterFunc(time.Second*1, func() {
			log.Println("Server started on port " + *portFlag)
		})

		if err := server.Start(); err != nil {
			log.Println(fmt.Sprintf("Failed to start HTTP server: %s", err.Error()))
			return
		}
	}
}
