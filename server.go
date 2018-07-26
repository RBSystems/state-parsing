package main

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/salt-translator-service/elk"
	"github.com/byuoitav/state-parsing/forwarding"
	"github.com/byuoitav/state-parsing/jobs"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	jobs.StartJobScheduler()

	go forwarding.StartDistributor()
	go forwarding.StartTicker(3000)

	port := ":10010"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	router.GET("/test", Status)

	router.PUT("/heartbeat", AddHeartbeat)
	router.PUT("/event", AddEvent)

	router.POST("/heartbeat", AddHeartbeat)
	router.POST("/event", AddEvent)

	router.PUT("/log-level/:level", log.SetLogLevel)
	router.GET("/log-level", log.GetLogLevel)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	err := router.StartServer(&server)
	if err != nil {
		log.L.Fatalf("error running server: %s", err)
	}
}

func Status(context echo.Context) error {
	return context.JSON(http.StatusOK, "Did you ever hear the tragedy of Darth Plagueis The Wise?")
}

func AddHeartbeat(context echo.Context) error {
	var heartbeat elk.Event
	err := context.Bind(&heartbeat)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid request body; not a valid event: %v", err))
	}

	jobs.HeartbeatChan <- heartbeat
	return context.JSON(http.StatusOK, "Success.")
}

func AddEvent(context echo.Context) error {
	var event elkreporting.ElkEvent
	err := context.Bind(&event)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid request body; not a valid event: %v", err))
	}

	jobs.EventChan <- event
	return context.JSON(http.StatusOK, "Success.")
}
