package statedefinition

import "time"

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
