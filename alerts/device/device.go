package device

import "github.com/byuoitav/state-parsing/eventforwarding"

type StaticDevice struct {
	Building              string           `json:"building,omitempty"`
	Control               string           `json:"control,omitempty"`
	Hostname              string           `json:"hostname,omitempty"`
	Room                  string           `json:"room,omitempty"`
	LastHeartbeat         string           `json:"last-heartbeat,omitempty"`
	Alerts                map[string]Alert `json:"alerts,omitempty"`
	SuppressNotifications string           `json:"suppress-notifications,omitempty"`
	Alerting              bool             `json:"alerting,omitempty"`
	Suppress              bool             `json:"suppress"`
	LastStateRecieved     string           `json:"last-state-recieved,omitempty"`
	ViewDashboard         string           `json:"view-dashboard,omitempty"`
	EnableNotifications   string           `json:"enable-notifications,omitempty"`
}

type Alert struct {
	Message   string `json:"message,omitempty"`
	AlertSent string `json:"alert-sent,omitempty"`
	Alerting  bool   `json:"alerting,omitempty"`
	Suppress  bool   `json:"Suppress,omitempty"`
}

type HeartbeatLostQueryResponse struct {
	Took     int  `json:"took,omitempty"`
	TimedOut bool `json:"timed_out,omitempty"`
	Shards   struct {
		Total      int `json:"total,omitempty"`
		Successful int `json:"successful,omitempty"`
		Skipped    int `json:"skipped,omitempty"`
		Failed     int `json:"failed,omitempty"`
	} `json:"_shards,omitempty"`
	Hits struct {
		Total    int     `json:"total,omitempty"`
		MaxScore float64 `json:"max_score,omitempty"`
		Hits     []struct {
			Index  string       `json:"_index,omitempty"`
			Type   string       `json:"_type,omitempty"`
			ID     string       `json:"_id,omitempty"`
			Score  float64      `json:"_score,omitempty"`
			Source StaticDevice `json:"_source,omitempty"`
		} `json:"hits,omitempty"`
	} `json:"hits,omitempty"`
}

//toMark is the list of rooms, There may be one or more of them
//secondaryAlertType is the type of alert marking as (e.g. heartbeat)
//secondarAlertData is the data to be filled there (e.g. last-heartbeat-received, etc)
func MarkGeneralAlerting(toMark []string, secondaryAlertType string, secondaryAlertData map[string]interface{}) {

	//build our general alerting
	alerting := eventforwarding.StateDistribution{
		Key:   "alerting",
		Value: true,
	}

	//bulid our specifc alert
	secondaryAlert := eventforwarding.StateDistribution{
		Key:   "alerts",
		Value: make(map[string]interface{}),
	}
	scondaryAlert[secondaryAlertType] = secondaryAlertData

	//ship it off to go with the rest
	for i := range toMark {
		eventforwarding.SendToStateBuffer(alerting, toMark[i], "device")
		eventforwarding.SendToStateBuffer(secondaryAlert, toMark[i], "device")
	}
}
