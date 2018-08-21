package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/state-parser/elk"
	"github.com/byuoitav/state-parser/jobs"
	"github.com/byuoitav/state-parser/state"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	go jobs.StartJobScheduler()
	go state.StartDistributor(3 * time.Second)

	port := ":10010"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	router.GET("/test", status)

	router.PUT("/heartbeat", addHeartbeat)
	router.PUT("/event", addEvent)
	router.POST("/heartbeat", addHeartbeat)
	router.POST("/event", addEvent)

	// dmps
	router.POST("/dmps/event", addDMPSEvent)
	router.POST("/dmps/heartbeat", addDMPSHeartbeat)

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
	var heartbeat events.Event
	err := context.Bind(&heartbeat)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid request body; not a valid heartbeat: %v", err))
	}
	log.L.Debugf("Received heartbeat: %+v", heartbeat)

	jobs.ProcessHeartbeat(heartbeat)
	return context.JSON(http.StatusOK, "Success.")
}

func addEvent(context echo.Context) error {
	var event elkreporting.ElkEvent
	err := context.Bind(&event)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid request body; not a valid event: %v", err))
	}
	log.L.Debugf("Received event: %+v", event)

	jobs.ProcessEvent(event)
	return context.JSON(http.StatusOK, "Success.")
}

func addDMPSEvent(context echo.Context) error {
	var event events.Event
	err := context.Bind(&event)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid request body; not a valid dmps event: %v", err))
	}
	log.L.Debugf("Received DMPS event: %+v", event)

	go state.Forward(event, elk.UpdateHeader{
		Index: elk.GenerateIndexName(elk.DMPS_EVENT),
		Type:  "dmpsevent",
	})
	return context.JSON(http.StatusOK, "Success.")
}

func addDMPSHeartbeat(context echo.Context) error {
	var event events.Event
	err := context.Bind(&event)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid request body; not a valid dmps event: %v", err))
	}
	log.L.Debugf("Received DMPS heartbeat: %+v", event)

	go state.Forward(event, elk.UpdateHeader{
		Index: elk.GenerateIndexName(elk.DMPS_HEARTBEAT),
		Type:  "dmpsheartbeat",
	})
	return context.JSON(http.StatusOK, "Success.")
}
