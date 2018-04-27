package device

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

type HeartbeatRestoredQueryResponse struct {
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
