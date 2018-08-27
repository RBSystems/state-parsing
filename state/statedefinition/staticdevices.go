package statedefinition

import (
	"time"

	"github.com/byuoitav/common/nerr"
)

//*************************
//IMPORTANT - if you add fields to this struct be sure to change the CompareDevices function
//*************************
type StaticDevice struct {
	//common fields
	ID                      string           `json:"ID,omitempty"`
	Alerting                *bool            `json:"alerting,omitempty"`
	Alerts                  map[string]Alert `json:"alerts,omitempty"`
	NotificationsSuppressed *bool            `json:"notifications-suppressed,omitempty"`
	Building                string           `json:"building,omitempty"`
	Room                    string           `json:"room,omitempty"`
	Hostname                string           `json:"hostname,omitempty"`
	LastStateReceived       time.Time        `json:"last-state-received,omitempty"`

	//semi-common fields
	LastHeartbeat time.Time `json:"last-heartbeat,omitempty"`
	LastUserInput time.Time `json:"last-user-input,omitempty"`
	Power         *bool     `json:"power,omitempty"`

	//Control Processor Specific Fields
	Websocket      string `json:"websocket,omitempty"`
	WebsocketCount *int   `json:"websocket-count,omitempty"`

	//Display Specific Fields
	Blanked *bool  `json:"blanked,omitempty"`
	Input   string `json:"input,omitempty"`

	//Audio Device Specific Fields
	Muted  *bool `json:"muted,omitempty"`
	Volume *int  `json:"volume,omitempty"`

	//Fields specific to Microphones
	BatteryChargeBars         *int   `json:"battery-charge-bars,omitempty"`
	BatteryChargeMinutes      *int   `json:"battery-charge-minutes,omitempty"`
	BatteryChargePercentage   *int   `json:"battery-charge-percentage,omitempty"`
	BatteryChargeHoursMinutes *int   `json:"battery-charge-hours-minutes,omitempty"`
	BatteryCycles             *int   `json:"battery-cycles,omitempty"`
	BatteryType               string `json:"battery-type,omitempty"`
	Interference              string `json:"*intererence,omitempty"`

	//meta fields for use in kibana
	Control               string `json:"control,omitempty"`                //the Hostname - used in a URL
	EnableNotifications   string `json:"enable-notifications,omitempty"`   //the Hostname - used in a URL
	SuppressNotifications string `json:"suppress-notifications,omitempty"` //the Hostname - used in a URL
	ViewDashboard         string `json:"ViewDashboard,omitempty"`          //the Hostname - used in a URL

}

//CompareDevices takes a base devices, and calculates the difference between the two, returning it in the staticDevice return value. Bool denotes if there were any differences
func CompareDevices(base, new StaticDevice) (diff StaticDevice, merged StaticDevice, changes bool, err *nerr.E) {

	//common fields
	diff.ID, merged.ID, changes = compareString(base.ID, new.ID, changes)
	diff.Alerting, merged.Alerting, changes = compareBool(base.Alerting, new.Alerting, changes)
	diff.Alerts, merged.Alerts, changes = compareAlerts(base.Alerts, new.Alerts, changes)
	diff.NotificationsSuppressed, merged.NotificationsSuppressed, changes = compareBool(base.NotificationsSuppressed, new.NotificationsSuppressed, changes)
	diff.Building, merged.Building, changes = compareString(base.Building, new.Building, changes)
	diff.Room, merged.Room, changes = compareString(base.Room, new.Room, changes)
	diff.Hostname, merged.Hostname, changes = compareString(base.Hostname, new.Hostname, changes)
	diff.LastStateReceived, merged.LastStateReceived, changes = compareTime(base.LastStateReceived, new.LastStateReceived, changes)

	//semi-common fields
	diff.LastHeartbeat, merged.LastHeartbeat, changes = compareTime(base.LastHeartbeat, new.LastHeartbeat, changes)
	diff.LastUserInput, merged.LastUserInput, changes = compareTime(base.LastUserInput, new.LastUserInput, changes)
	diff.Power, merged.Power, changes = compareBool(base.Power, new.Power, changes)

	//Conrol processor specific fields
	diff.Websocket, merged.Websocket, changes = compareString(base.Websocket, new.Websocket, changes)
	diff.WebsocketCount, merged.WebsocketCount, changes = compareInt(base.WebsocketCount, new.WebsocketCount, changes)

	//Display specific fields
	diff.Blanked, merged.Blanked, changes = compareBool(base.Blanked, new.Blanked, changes)
	diff.Input, merged.Input, changes = compareString(base.Input, new.Input, changes)

	//Audio Device specific fields
	diff.Muted, merged.Muted, changes = compareBool(base.Muted, new.Muted, changes)
	diff.Volume, merged.Volume, changes = compareInt(base.Volume, new.Volume, changes)

	//Microphone specific fields
	diff.BatteryChargeBars, merged.BatteryChargeBars, changes = compareInt(base.BatteryChargeBars, new.BatteryChargeBars, changes)
	diff.BatteryChargeMinutes, merged.BatteryChargeMinutes, changes = compareInt(base.BatteryChargeMinutes, new.BatteryChargeMinutes, changes)
	diff.BatteryChargePercentage, merged.BatteryChargePercentage, changes = compareInt(base.BatteryChargePercentage, new.BatteryChargePercentage, changes)
	diff.BatteryChargeHoursMinutes, merged.BatteryChargeHoursMinutes, changes = compareInt(base.BatteryChargeHoursMinutes, new.BatteryChargeHoursMinutes, changes)
	diff.BatteryCycles, merged.BatteryCycles, changes = compareInt(base.BatteryCycles, new.BatteryCycles, changes)
	diff.BatteryType, merged.BatteryType, changes = compareString(base.BatteryType, new.BatteryType, changes)
	diff.Interference, merged.Interference, changes = compareString(base.Interference, new.Interference, changes)

	//meta fields
	diff.Control, merged.Control, changes = compareString(base.Control, new.Control, changes)
	diff.EnableNotifications, merged.EnableNotifications, changes = compareString(base.EnableNotifications, new.EnableNotifications, changes)
	diff.SuppressNotifications, merged.SuppressNotifications, changes = compareString(base.SuppressNotifications, new.SuppressNotifications, changes)
	diff.ViewDashboard, merged.ViewDashboard, changes = compareString(base.ViewDashboard, new.ViewDashboard, changes)

	return
}

func compareString(base, new string, changes bool) (string, string, bool) {
	if new != "" {
		if base != new {
			return new, new, true
		}
	}
	return "", base, false || changes
}

func compareBool(base, new *bool, changes bool) (*bool, *bool, bool) {
	if new != nil {
		if *base != *new {
			return new, new, true
		}
	}
	return nil, base, false || changes
}

func compareInt(base, new *int, changes bool) (*int, *int, bool) {
	if new != nil {
		if *base != *new {
			return new, new, true
		}
	}
	return nil, base, false || changes
}

func compareTime(base, new time.Time, changes bool) (time.Time, time.Time, bool) {
	if !new.IsZero() {
		if !new.Equal(base) {
			return new, new, true
		}
	}
	return time.Time{}, base, false || changes
}
