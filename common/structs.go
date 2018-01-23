package common

type UpdateHeader struct {
	ID    string `json:"_id"`
	Type  string `json:"_type"`
	Index string `json:"_index"`
}

type UpdateBody struct {
	Doc    map[string]string `json:"doc"`
	Upsert bool              `json:"doc_as_upsert"`
}
