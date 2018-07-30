package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/salt-translator-service/elk"
	"github.com/byuoitav/state-parser/forwarding"
	"github.com/byuoitav/state-parser/jobs"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	jobs.StartJobScheduler()
	go forwarding.StartDistributor(3 * time.Second)

	//	go forwarding.StartDistributor()
	//	go forwarding.StartTicker(3 * time.Second)

	port := ":10010"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	router.GET("/test", status)

	router.PUT("/heartbeat", addHeartbeat)
	router.PUT("/event", addEvent)

	router.POST("/heartbeat", addHeartbeat)
	router.POST("/event", addEvent)

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

func status(context echo.Context) error {
	return context.JSON(http.StatusOK, "Did you ever hear the tragedy of Darth Plagueis The Wise?")
}

func addHeartbeat(context echo.Context) error {
	var heartbeat elk.Event
	err := context.Bind(&heartbeat)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid request body; not a valid event: %v", err))
	}
	log.L.Debugf("Received heartbeat: %+v", heartbeat)

	// forward event
	forwarding.Forward(heartbeat, jobs.HeartbeatForward)

	jobs.HeartbeatChan <- heartbeat
	return context.JSON(http.StatusOK, "Success.")
}

func addEvent(context echo.Context) error {
	var event elkreporting.ElkEvent
	err := context.Bind(&event)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid request body; not a valid event: %v", err))
	}
	log.L.Debugf("Received event: %+v", event)

	// forward event
	forwarding.Forward(event, jobs.APIForward)

	jobs.EventChan <- event
	return context.JSON(http.StatusOK, "Success.")
}
