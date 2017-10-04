package eventforwarding

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/event-translator-microservice/elkreporting"
	heartbeat "github.com/byuoitav/salt-translator-service/elk"
	"github.com/labstack/echo"
)

var eventIngestionChannel chan elkreporting.ElkEvent
var heartbeatIngestionChannel chan heartbeat.Event

func AddEvent(context echo.Context) error {
	//take the event that was sent and ingest the event down the ingestion channel

	var event elkreporting.ElkEvent
	err := context.Bind(&event)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Error with the body, not a valid event: %v: ", err.Error()))
	}

	eventIngestionChannel <- event
	return context.JSON(http.StatusOK, "Success.")
}

func AddHeartbeat(context echo.Context) error {
	//take the event that was sent and ingest the event down the ingestion channel

	var event heartbeat.Event
	err := context.Bind(&event)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Sprintf("Error with the body, not a valid event: %v: ", err.Error()))
	}

	heartbeatIngestionChannel <- event
	return context.JSON(http.StatusOK, "Success.")
}

func Test(context echo.Context) error {
	return context.JSON(http.StatusOK, "Did you ever hear the tragedy of Darth Plagueis The Wise?")
}
