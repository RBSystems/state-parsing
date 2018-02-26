package base

//Alerts are the types that get generated. For now there's only Slack alerts, the intention is to add at least E-Mail alerts and MOM alerts
type Alert struct {
	AlertType string //The type of alert, see constants.go to see the available values
	Content   []byte //The content of the alert to send
	Device    string //the Device the alert corresponds to
}

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

//AlertReport report on the success of alert notifications
type AlertReport struct {
	Alert
	Success bool
	Message string
}
