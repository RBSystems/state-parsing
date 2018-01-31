package room

import "github.com/byuoitav/state-parsing/alerts/device"

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
