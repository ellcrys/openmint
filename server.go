package main 

import (
	"flag"
	"log"
	"time"
	"fmt"

	"github.com/ellcrys/util"
	"github.com/ellcrys/openmint/www"
	"github.com/labstack/echo/engine/standard"
)

func main() {

	// determine appropriate port number
	var portEnv  = util.Env("PORT", "3001")
	var runSeed	 = flag.Bool("seed", false, "Seed the application")
	var portFlag = flag.String("port", portEnv, "set port. Default: " + portEnv)
	flag.Parse()

	if flag.Parsed() {

		// create new router
		router := www.App(false, *runSeed)
		server := standard.New(":" + *portFlag)
		server.SetHandler(router)
		
		time.AfterFunc(time.Second * 1, func () {
			log.Println("Server started on port " + *portFlag)	
		})
		
		if err := server.Start(); err != nil {
			log.Println(fmt.Sprintf("Failed to start HTTP server: %s", err.Error()))
			return
		}
	}
}	