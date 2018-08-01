package marking

import "github.com/byuoitav/state-parser/forwarding"

func MarkRoomGeneralAlerting(toMark []string, alerting bool) {
	//build our state
	state := forwarding.State{
		Key:   "alerting",
		Value: alerting,
	}

	//ship it off to go with the rest
	for i := range toMark {
		state.ID = toMark[i]
		forwarding.BufferState(state, "room")
	}
}
