package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/shure-audio-microservice/handlers"
	"github.com/byuoitav/shure-audio-microservice/publishing"
	"github.com/byuoitav/shure-audio-microservice/reporting"
	"github.com/fatih/color"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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

	hostname := os.Getenv("PI_HOSTNAME")
	building := strings.Split(hostname, "-")[0]
	room := strings.Split(hostname, "-")[1]
	log.L.Infof("%s", color.HiBlueString("[server] detected hostname: %s", hostname))

	if strings.EqualFold(strings.Split(hostname, "-")[2], "CP1") {
		//start live monitoring/publishing
		go reporting.Monitor(building, room)
	}

	port := ":8013"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	// Use the `secure` routing group to require authentication
	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	//TODO share one connection!

	secure.GET("/health", handlers.Health)

	secure.PUT("/raw", handlers.Raw)

	secure.GET("/:address/:channel/battery/:format", handlers.Battery)

	secure.GET("/:address/:channel/power/status", handlers.Power)

	secure.PUT("/log-level/:level", log.SetLogLevel)
	secure.GET("/log-level", log.GetLogLevel)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)
}
