package elk

type UpdateHeader struct {
	ID    string `json:"_id,omitempty"`
	Type  string `json:"_type,omitemtpy"`
	Index string `json:"_index,omitempty"`
}

type DeviceUpdateInfo struct {
	Info string `json:"Info"`
	Name string `json:"Name"`
}

type UpdateBody struct {
	Doc    map[string]interface{} `json:"doc"`
	Upsert bool                   `json:"doc_as_upsert"`
}

type StaticDevice struct {
	Building              string           `json:"building,omitempty"`
	Control               string           `json:"control,omitempty"`
	Hostname              string           `json:"hostname,omitempty"`
	Room                  string           `json:"room,omitempty"`
	LastHeartbeat         string           `json:"last-heartbeat,omitempty"`
	Alerts                map[string]Alert `json:"alerts,omitempty"`
	SuppressNotifications string           `json:"suppress-notifications,omitempty"`
	Alerting              bool             `json:"alerting,omitempty"`
	Suppress              bool             `json:"notifications-suppressed"`
	LastStateRecieved     string           `json:"last-state-recieved,omitempty"`
	ViewDashboard         string           `json:"view-dashboard,omitempty"`
	EnableNotifications   string           `json:"enable-notifications,omitempty"`
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
	Alerts            map[string]Alert
	Suppressed        bool `json:"notifications-suppressed"`
}

type Alert struct {
	Message   string `json:"message,omitempty"`
	AlertSent string `json:"alert-sent,omitempty"`
	Alerting  bool   `json:"alerting,omitempty"`
	Suppress  bool   `json:"Suppress,omitempty"`
}
