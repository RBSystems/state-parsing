package room

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/alerts/device"
	"github.com/byuoitav/state-parsing/eventforwarding"
	"github.com/fatih/color"
)

type RoomQueryResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total    int     `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index  string     `json:"_index"`
			Type   string     `json:"_type"`
			ID     string     `json:"_id"`
			Score  float64    `json:"_score"`
			Source StaticRoom `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type StaticRoom struct {
	Building          string `json:"building"`
	EnableAlerts      string `json:"enable-alerts"`
	LastStateRecieved string `json:"last-state-recieved"`
	LastUserInput     string `json:"last-user-input"`
	Room              string `json:"room"`
	SuspendAlerts     string `json:"suspend-alerts"`
	ViewAlerts        string `json:"view-alerts"`
	ViewDevices       string `json:"view-devices"`
	Power             string `json:"power"`
	Alerting          bool   `json:"alerting"`
	Alerts            map[string]device.Alert
	Suppressed        bool `json:"suppressed"`
}

type StatiRoomWrapper struct {
	Index string `json:"_index"`
	Type  string `json:"_type"`
	ID    string `json:"_id"`
}

func MarkGeneralAlerting(toMark []string) {

	//build our state
	alerting := eventforwarding.StateDistribution{
		Key:   "alerting",
		Value: true,
	}

	//ship it off to go with the rest
	for i := range toMark {
		eventforwarding.SendToStateBuffer(alerting, toMark[i], "room")
	}
}

func GetRoomsBulk(rooms []string) ([]StaticRoom, error) {

	//assume that the rooms is the array of ID's
	query := base.IDQuery{}
	query.Query.IDS.Type = "room"
	query.Query.IDS.Values = rooms

	b, err := json.Marshal(&query)
	if err != nil {
		log.Printf(color.HiRedString("Error: Could not marshal the json: %v", err.Error()))
		return []StaticRoom{}, err
	}

	url := fmt.Sprintf("%v/%v/_search", os.Getenv("ELK_ADDR"), os.Getenv("ELK_STATIC_ROOM_INDEX"))
	log.Printf("Body: %s", b)

	respCode, body, err := base.MakeELKRequest(url, "POST", b, 1)
	if err != nil {
		log.Printf(color.HiRedString("Error making the request: %v", err.Error()))
		return []StaticRoom{}, err

	}

	if respCode/100 != 2 {
		log.Printf(color.HiRedString("Non 200 response recieved: %v. Body: %s", respCode, body))
		return []StaticRoom{}, errors.New(fmt.Sprintf("Non 200 response recieved: %v. Body: %s", respCode, body))
	}

	//we have the body, unmarshal it

	rresp := RoomQueryResponse{}
	err = json.Unmarshal(body, &rresp)
	if err != nil {
		log.Printf(color.HiRedString("Could not unmarshal response: %v", err.Error()))
		return []StaticRoom{}, err
	}

	toReturn := []StaticRoom{}

	for i := range rresp.Hits.Hits {
		toReturn = append(toReturn, rresp.Hits.Hits[i].Source)
	}

	return toReturn, nil
}
