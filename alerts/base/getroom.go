package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/byuoitav/state-parsing/alerts/room"
	"github.com/fatih/color"
)

type idsQuery struct {
	Query struct {
		IDS struct {
			Type   string   `json:"type"`
			Values []string `json:"values"`
		} `json:"ids"`
	} `json:"query"`
}

func GetRoomsBulk(rooms []string) ([]room.StaticRoom, error) {

	//assume that the rooms is the array of ID's
	query := idsQuery{}
	query.Query.IDS.Type = "room"
	query.Query.IDS.Values = rooms

	b, err := json.Marshal(&query)
	if err != nil {
		log.Printf(color.HiRedString("Error: Could not marshal the json: %v", err.Error()))
		return []room.StaticRoom{}, err
	}

	url := fmt.Sprintf("%v/%v/_search", os.Getenv("ELK_ADDR"), os.Getenv("ELK_STATIC_ROOM_INDEX"))
	log.Printf("Body: %s", b)

	respCode, body, err := MakeELKRequest(url, "POST", b, 1)
	if err != nil {
		log.Printf(color.HiRedString("Error making the request: %v", err.Error()))
		return []room.StaticRoom{}, err

	}

	if respCode/100 != 2 {
		log.Printf(color.HiRedString("Non 200 response recieved: %v. Body: %s", respCode, body))
		return []room.StaticRoom{}, errors.New(fmt.Sprintf("Non 200 response recieved: %v. Body: %s", respCode, body))
	}

	//we have the body, unmarshal it

	rresp := room.RoomQueryResponse{}
	err = json.Unmarshal(body, &rresp)
	if err != nil {
		log.Printf(color.HiRedString("Could not unmarshal response: %v", err.Error()))
		return []room.StaticRoom{}, err
	}

	toReturn := []room.StaticRoom{}

	for i := range rresp.Hits.Hits {
		toReturn = append(toReturn, rresp.Hits.Hits[i].Source)
	}

	return toReturn, nil
}
