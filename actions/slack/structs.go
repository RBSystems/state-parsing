package slack

type SlackAlert struct {
	Attachments []SlackAttachment `json:"attachments,omitempty"`
	Markdown    bool              `json:"mrkdwn"`
	Text        string            `json:"text, omitempty"`
}

type SlackAttachment struct {
	Fallback  string            `json:"fallback, omitempty"`
	Pretext   string            `json:"pretext, omitempty"`
	Title     string            `json:"title, omitempty"`
	TitleLink string            `json:"title_link, omitempty"`
	Text      string            `json:"text, omitempty"`
	Color     string            `json:"color, omitempty"`
	Fields    []SlackAlertField `json:"fields,omitempty"`
}

type SlackAlertField struct {
	Title string `json:"title,omitemtpy"`
	Value string `json:"value,omitemtpy"`
	Short bool   `json:"short,omitemtpy"`
}
