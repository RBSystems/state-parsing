package marking

import "github.com/byuoitav/state-parser/state"

func MarkRoomGeneralAlerting(toMark []string, alerting bool) {
	//build our state
	val := state.State{
		Key:   "alerting",
		Value: alerting,
	}

	//ship it off to go with the rest
	for i := range toMark {
		val.ID = toMark[i]
		state.BufferState(val, state.ROOM)
	}
}
