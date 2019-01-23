package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/byuoitav/common"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/v2/auth"
	"github.com/byuoitav/shure-audio-microservice/handlers"
	"github.com/byuoitav/shure-audio-microservice/publishing"
	"github.com/byuoitav/shure-audio-microservice/reporting"
	"github.com/fatih/color"
)

/**
STATES TO QUERY:

	battery bars

	battery time

	battery type

	power status

STATES TO REPORT

	power on

	power off

	RF interferance
**/

const PORT = 2202

func main() {

	log.L.Infof("%s", color.HiGreenString("[server] starting Shure Audio Microservice..."))

	//request event router subsribe to events
	go publishing.Start()

	hostname := os.Getenv("SYSTEM_ID")
	building := strings.Split(hostname, "-")[0]
	room := strings.Split(hostname, "-")[1]
	log.L.Infof("%s", color.HiBlueString("[server] detected hostname: %s", hostname))

	//we only want to monitor if we're the first device in the room
	if strings.EqualFold(strings.Split(hostname, "-")[2], "CP1") {
		//start live monitoring/publishing
		go reporting.Monitor(building, room)
	}

	port := ":8013"

	//TODO share one connection!

	router := common.NewRouter()

	write := router.Group("", auth.AuthorizeRequest("write-state", "room", auth.LookupResourceFromAddress))
	write.PUT("/raw", handlers.Raw)

	read := router.Group("", auth.AuthorizeRequest("read-state", "room", auth.LookupResourceFromAddress))
	read.GET("/:address/:channel/battery/:format", handlers.Battery)
	read.GET("/:address/:channel/power/status", handlers.Power)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)
}
