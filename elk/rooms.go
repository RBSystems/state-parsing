package elk

import (
	"encoding/json"
	"fmt"

	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/state-parser/state/statedefinition"
)

type roomQueryResponse struct {
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
			Index  string                     `json:"_index"`
			Type   string                     `json:"_type"`
			ID     string                     `json:"_id"`
			Score  float64                    `json:"_score"`
			Source statedefinition.StaticRoom `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func GetRoomsBulk(rooms []string) ([]statedefinition.StaticRoom, *nerr.E) {
	//assume that the rooms is the array of ID's
	query := IDQuery{}
	query.Query.IDS.Type = "room"
	query.Query.IDS.Values = rooms

	endpoint := fmt.Sprintf("/%s/_search", "oit-av-static-rooms")
	body, err := MakeGenericELKRequest("POST", endpoint, query)
	if err != nil {
		return []statedefinition.StaticRoom{}, err.Addf("failed to get rooms bulk")
	}

	//we have the body, unmarshal it
	rresp := roomQueryResponse{}
	gerr := json.Unmarshal(body, &rresp)
	if err != nil {
		return []statedefinition.StaticRoom{}, nerr.Translate(gerr).Addf("failed to get rooms bulk")
	}

	toReturn := []statedefinition.StaticRoom{}
	for i := range rresp.Hits.Hits {
		toReturn = append(toReturn, rresp.Hits.Hits[i].Source)
	}

	return toReturn, nil
}

func AlertingSuppressedRooms(toCheck []statedefinition.StaticRoom) (map[string]bool, map[string]bool) {
	alerting := make(map[string]bool)
	suppressed := make(map[string]bool)
	//go through each room in the array and check if it's already alerting

	for i := range toCheck {
		alerting[toCheck[i].RoomID] = *toCheck[i].Alerting
		suppressed[toCheck[i].RoomID] = *toCheck[i].NotificationsSuppressed
	}

	return alerting, suppressed
}
