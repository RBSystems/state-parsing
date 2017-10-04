package main

import (
	"net/http"

	"github.com/byuoitav/state-parsing/eventforwarding"
	"github.com/byuoitav/state-parsing/jobs"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	o := jobs.Orchestrator{}
	o.Start()

	go eventforwarding.StartDistributor()
	go eventforwarding.StartTicker(3000)
	go eventforwarding.Init()

	port := ":10010"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	router.GET("/test", eventforwarding.Test)

	router.PUT("/heartbeat", eventforwarding.AddHeartbeat)
	router.PUT("/event", eventforwarding.AddEvent)

	router.POST("/heartbeat", eventforwarding.AddHeartbeat)
	router.POST("/event", eventforwarding.AddEvent)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)
}
