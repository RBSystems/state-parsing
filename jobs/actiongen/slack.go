package actiongen

import (
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/state-parser/actions/action"
)

//GenSlackAction .
func GenSlackAction(config Config, event events.Event, device string) (action.Payload, *nerr.E) {
	return action.Payload{}, nil
}
