package statedefinition

import "time"

type StaticDevice struct {
	//common fields
	ID                      string           `json:"ID,omitempty"`
	Alerting                bool             `json:"alerting,omitempty"`
	Alerts                  map[string]Alert `json:"alerts,omitempty"`
	NotificationsSuppressed bool             `json:"notifications-suppressed,omitempty"`
	Building                string           `json:"building,omitempty"`
	Room                    string           `json:"room,omitempty"`
	Hostname                string           `json:"hostname,omitempty"`
	LastStateRecieved       time.Time        `json:"last-state-received,omitempty"`

	//semi-common fields
	LastHeartbeat time.Time `json:"last-heartbeat,omitempty"`
	LastUserInput time.Time `json:"last-user-input,omitempty"`
	Power         bool      `json:"power,omitempty"`

	//Control Processor Specific Fields
	Websocket      string `json:"websocket,omitempty"`
	WebsocketCount int    `json:"websocket-count,omitempty"`

	//Display Specific Fields
	Blanked bool   `json:"blanked,omitempty"`
	Input   string `json:"input,omitempty"`

	//Audio Device Specific Fields
	Muted  bool `json:"muted,omitempty"`
	Volume int  `json:"volume,omitempty"`

	//Fields specific to Microphones
	BatteryChargeBars        int    `json:"battery-charge-bars,omitempty"`
	BatteryChargeMinutes     int    `json:"battery-charge-minutes,omitempty"`
	BatteryChargePercentage  int    `json:"battery-charge-percentage,omitempty"`
	BatteryLevelHoursMinutes int    `json:"battery-charge-hours-minutes,omitempty"`
	BatteryCycles            int    `json:"battery-cycles,omitempty"`
	BatteryType              string `json:"battery-type,omitempty"`
	Interference             string `json:"intererence,omitempty"`

	//meta fields for use in kibana
	Control              string `json:"control,omitempty"`                //the Hostname - used in a URL
	EnableNotification   string `json:"enable-notifications,omitempty"`   //the Hostname - used in a URL
	SuppressNotification string `json:"suppress-notifications,omitempty"` //the Hostname - used in a URL
	ViewDashboard        string `json:"ViewDashboard,omitempty"`          //the Hostname - used in a URL

}

type Alert struct {
	AlertSent time.Time `json:"alert-sent,omitempty"`
	Alerting  bool      `json:"alerting,omitempty"`
	Message   string    `json:"message,omitempty"`
}

type StaticRoom struct {
	//information fields
	Building string `json:"building,omitempty"`
	Room     string `json:"room,omitempty"`

	//State fields
	NotificationsSuppressed bool `json:"notifications-suppressed,omitempty"`
	Alerting                bool `json:"alerting,omitempty"`

	LastStateRecieved time.Time `json:"last-state-received,omitempty"`
	LastHeartbeat     time.Time `json:"last-heartbeat,omitempty"`
	LastUserInput     time.Time `json:"last-user-input,omitempty"`

	Power bool `json:"power,omitempty"`

	//meta fields for Kibana
	ViewDevices          string `json:"view-devices"`
	ViewAlerts           string `json:"view-alerts"`
	EnableNotification   string `json:"enable-notifications,omitempty"`   //the Hostname - used in a URL
	SuppressNotification string `json:"suppress-notifications,omitempty"` //the Hostname - used in a URL
}
