package elk

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
