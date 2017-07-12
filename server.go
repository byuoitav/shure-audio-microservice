package main

import (
	"net/http"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/av-api/health"
	"github.com/byuoitav/shure-audio-microservice/handlers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {

	port := ":8013"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	// Use the `secure` routing group to require authentication
	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	//TODO add endpoints!

	secure.GET("/health", handlers.Health)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	go health.StartupCheckAndReport()

	router.StartServer(&server)
}
