package actions

import "github.com/byuoitav/state-parsing/actions/action"

var ingestionMap map[string]chan action.Action

type actionManager struct {
	Name       string
	Action     Action
	ActionChan chan action.Action
}

func StartActionScheduler() {
	// build each of the individual action managers
	for name, act := range Actions {
		ingestionMap[name] = make(chan action.Action, 1000)

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
	// TODO scale number of action managers as size of payload chan increases
	for action := range a.ActionChan {
		a.Action.Execute(action)
	}
}
