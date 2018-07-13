package marking

import "github.com/byuoitav/state-parsing/forwarding"

func MarkRoomGeneralAlerting(toMark []string, alerting bool) {
	//build our state
	state := forwarding.StateDistribution{
		Key:   "alerting",
		Value: alerting,
	}

	//ship it off to go with the rest
	for i := range toMark {
		forwarding.SendToStateBuffer(state, toMark[i], "room")
	}
}
