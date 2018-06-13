package generalalert

import (
	"github.com/byuoitav/state-parsing/logger"
	"github.com/byuoitav/state-parsing/tasks/names"
	"github.com/byuoitav/state-parsing/update"
)

type GeneralAlertUpdater struct {
	update.Updater
}

func (r *GeneralAlertUpdater) Init() {
	r.Logger = logger.New(names.GENERAL_ALERT_UPDATE, logger.INFO)
}

func (r *RoomUpdater) Run() error {
	body, err := GeneralAlertQuery.MakeELKRequest(r.LogLevel, r.Name)
	if err != nil {
		r.Error("error with the initial query: %s", err)
		return err
	}

}
