package generalalert

import (
	"encoding/json"

	"github.com/byuoitav/state-parsing/eventforwarding"
	"github.com/byuoitav/state-parsing/logger"
	"github.com/byuoitav/state-parsing/tasks/names"
	"github.com/byuoitav/state-parsing/update"
)

type GeneralAlertUpdater struct {
	update.Updater
}

func (r *GeneralAlertUpdater) Init() {
	r.Logger = logger.New(names.GENERAL_ALERT_UPDATE, logger.VERBOSE)
}

func (r *GeneralAlertUpdater) Run() error {
	r.Logger.Verbose("Starting run of general alert clearing")

	//the query is constructed such that only elements that have a general alerting set to true, but no specific alersts return.
	body, err := GeneralAlertQuery.MakeELKRequest(r.LogLevel, r.Name)
	if err != nil {
		r.Error("error with the initial query: %s", err)
		return err
	}

	var resp GeneralAlertQueryResponse

	err = json.Unmarshal(body, &resp)
	if err != nil {
		r.Error("couldn't unmarshal rsponse: %s", body)
		return err
	}

	r.Logger.Verbose("Query is back. Starting processing")

	alertcleared := eventforwarding.StateDistribution{
		Key:   "alerting",
		Value: false,
	}

	//go through and mark each of these rooms as not alerting, in the general
	for _, hit := range resp.Hits.Hits {
		r.Logger.Verbose("running for %v", hit.ID)
		eventforwarding.SendToStateBuffer(alertcleared, hit.ID, "device")
	}

	return nil
}
