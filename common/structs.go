package common

type UpdateHeader struct {
	ID    string `json:"_id"`
	Type  string `json:"_type"`
	Index string `json:"_index"`
}

type DeviceUpdateInfo struct {
	Info string `json:"Info"`
	Name string `json:"Name"`
}

type UpdateBody struct {
	Doc    map[string]interface{} `json:"doc"`
	Upsert bool                   `json:"doc_as_upsert"`
}
