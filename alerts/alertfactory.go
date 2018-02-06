package alerts

import (
	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/byuoitav/state-parsing/alerts/heartbeat"
)

var alertFactories = map[string]AlertFactory{}

//AlertFactory corresponds to a struct that is run to generate alerts.
type AlertFactory interface {
	Run(loggingLevel int) (map[string][]base.Alert, error)
}

//returns the alert factory, the bool indicates if it was a valid name
func GetAlertFactory(name string) (AlertFactory, bool) {
	if len(alertFactories) == 0 {
		alertFactories = make(map[string]AlertFactory)

		//add the factories here

		alertFactories[base.LOST_HEARTBEAT] = &heartbeat.LostHeartbeatAlertFactory{}
	}

	v, ok := alertFactories[name]
	return v, ok

}
