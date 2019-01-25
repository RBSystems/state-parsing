package main

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/common/log"
	v2 "github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/state-parser/config"
	"github.com/byuoitav/state-parser/jobs"
	"github.com/byuoitav/state-parser/state/cache"
	"github.com/byuoitav/state-parser/state/forwarding"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	echopprof "github.com/sevenNt/echo-pprof"
)

func main() {
	log.SetLevel("warn")

	go jobs.StartJobScheduler()

	c := config.GetConfig()

	pre, _ := log.GetLevel()

	log.SetLevel("info")
	log.L.Infof("Initializing Caches")
	cache.InitializeCaches(c.Caches, forwarding.GetManagersForType)
	log.L.Infof("Caches Initialized.")
	log.SetLevel(pre)

	port := ":10011"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	router.GET("/test", status)
	router.GET("/cachestatus", cacheStatus)
	router.GET("/queuestatus", queueStatus)

	router.POST("/v2/event", addV2Event)
	router.POST("/legacy/v2/event", addV2LegacyEvent)

	router.PUT("/log-level/:level", log.SetLogLevel)
	router.GET("/log-level", log.GetLogLevel)

	echopprof.Wrap(router)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	err := router.StartServer(&server)
	if err != nil {
		log.L.Fatalf("error running server: %s", err)
	}
}

//CacheStatusStruct .
type CacheStatusStruct struct {
	DeviceCount int
	DeviceList  []string
}

func queueStatus(context echo.Context) error {
	return context.JSON(http.StatusOK, jobs.GetQueueSize())
}

func cacheStatus(context echo.Context) error {
	toReturn := map[string]CacheStatusStruct{}

	config := config.GetConfig()

	for _, ca := range config.Caches {
		c := cache.GetCache(ca.CacheType)
		//asser it's a mem
		count, l, _ := c.GetDeviceManagerList()
		toReturn[ca.CacheType] = CacheStatusStruct{
			DeviceList:  l,
			DeviceCount: count,
		}

	}

	return context.JSON(http.StatusOK, toReturn)
}

func status(context echo.Context) error {
	return context.JSON(http.StatusOK, "Did you ever hear the tragedy of Darth Plagueis The Wise?")
}

func addV2Event(context echo.Context) error {
	var event v2.Event
	err := context.Bind(&event)
	if err != nil {
		log.L.Debugf("Bad event: %v", err.Error())
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
