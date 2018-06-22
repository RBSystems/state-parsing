package base

//Alerts are the types that get generated. For now there's only Slack alerts, the intention is to add at least E-Mail alerts and MOM alerts
type Alert struct {
	AlertType string //The type of alert, see constants.go to see the available values
	Content   []byte //The content of the alert to send
	Device    string //the Device the alert corresponds to
}

//AlertReport report on the success of alert notifications
type AlertReport struct {
	Alert
	Success bool
	Message string
}
