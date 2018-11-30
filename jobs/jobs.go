package jobs

import (
	"github.com/byuoitav/state-parser/actions/action"
	"github.com/byuoitav/state-parser/config"
	"github.com/byuoitav/state-parser/jobs/eventbased"
	"github.com/byuoitav/state-parser/jobs/timebased"
	"github.com/byuoitav/state-parser/jobs/timebased/statequery"
)

// Job .
type Job interface {
	Run(ctx config.JobInputContext, actionWrite chan action.Payload)
	GetName() string
}

// Jobs .
var Jobs = map[string]Job{
	timebased.RoomUpdate:           &timebased.RoomUpdateJob{},
	timebased.GeneralAlertClearing: &timebased.GeneralAlertClearingJob{},
	eventbased.SimpleForwarding:    &eventbased.SimpleForwardingJob{},
	eventbased.GenAction:           &eventbased.GenActionJob{},
	statequery.StateQuery:          &statequery.QueryJob{},
}
