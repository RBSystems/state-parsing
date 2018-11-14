package main

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/log"
	v2 "github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/state-parser/jobs"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	log.SetLevel("debug")
	go jobs.StartJobScheduler()

	port := ":10011"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	router.GET("/test", status)

	router.PUT("/heartbeat", addHeartbeat)
	router.PUT("/event", addEvent)
	router.POST("/heartbeat", addHeartbeat)
	router.POST("/event", addEvent)

	router.POST("/v2/event", addV2Event)
	router.POST("/legacy/v2/event", addV2LegacyEvent)

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
	var heartbeat elkreporting.ElkEvent
	err := context.Bind(&heartbeat)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid request body; not a valid heartbeat: %v", err))
	}
	log.L.Debugf("Received heartbeat: %+v", heartbeat)

	jobs.ProcessEvent(heartbeat)
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

func addV2Event(context echo.Context) error {
	var event v2.Event
	err := context.Bind(&event)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid request body; not a valid event: %v", err))
	}
	log.L.Debugf("Received event: %+v", event)

	jobs.ProcessV2Event(event)
	return context.JSON(http.StatusOK, "Success.")
}

func addV2LegacyEvent(context echo.Context) error {
	var event v2.Event
	err := context.Bind(&event)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid request body; not a valid event: %v", err))
	}
	log.L.Debugf("Received event: %+v", event)

	jobs.ProcessLegacyV2Event(event)
	return context.JSON(http.StatusOK, "Success.")
}

func addDMPSEvent(context echo.Context) error {
	var event events.Event
	err := context.Bind(&event)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid request body; not a valid dmps event: %v", err))
	}
	log.L.Debugf("Received DMPS event: %+v", event)
	/*
		go state.Forward(event, elk.UpdateHeader{
			Index: elk.GenerateIndexName(elk.DMPS_EVENT),
			Type:  "dmpsevent",
		})
	*/
	return context.JSON(http.StatusOK, "Success.")
}

func addDMPSHeartbeat(context echo.Context) error {
	var event events.Event
	err := context.Bind(&event)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid request body; not a valid dmps event: %v", err))
	}
	log.L.Debugf("Received DMPS heartbeat: %+v", event)

	/*
		go state.Forward(event, elk.UpdateHeader{
			Index: elk.GenerateIndexName(elk.DMPS_HEARTBEAT),
			Type:  "dmpsheartbeat",
		})
	*/
	return context.JSON(http.StatusOK, "Success.")
}
