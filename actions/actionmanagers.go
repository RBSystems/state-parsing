package actions

import (
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parsing/actions/action"
)

var ingestionMap map[string]chan action.Action

type actionManager struct {
	Name       string
	Action     Action
	ActionChan chan action.Action
}

func init() {
	ingestionMap = make(map[string]chan action.Action)
}

func StartActionManagers() {
	var actionList []string
	for name, _ := range Actions {
		actionList = append(actionList, name)
	}

	log.L.Infof("Starting action scheduler. Executing action types: %v", actionList)

	// build each of the individual action managers
	for name, act := range Actions {
		ingestionMap[name] = make(chan action.Action, 2000) // TODO make this size configurable?

		manager := &actionManager{
			Name:       name,
			Action:     act,
			ActionChan: ingestionMap[name],
		}
		go manager.start()
	}
}

func Execute(actions []action.Action) {
	if len(actions) == 0 {
		return
	}

	for _, action := range actions {
		if _, ok := ingestionMap[action.Type]; ok {
			ingestionMap[action.Type] <- action
		}
	}
}

func (a *actionManager) start() {
	// TODO scale number of action managers as size of payload chan increases?
	for act := range a.ActionChan {
		go func(action action.Action) {
			result := a.Action.Execute(action)
			if result.Error != nil {
				log.L.Warnf("failed to execute %s action: %s", result.Action.Type, result.Error.String())
			}
		}(act)
	}
}
