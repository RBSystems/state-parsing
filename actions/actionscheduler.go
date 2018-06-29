package actions

var ingestionMap map[string]chan ActionPayload

type actionManager struct {
	Name        string
	Action      Action
	PayloadChan chan ActionPayload
}

func StartActionScheduler() {
	// build each of the individual action managers
	for name, action := range Actions {
		ingestionMap[name] = make(chan ActionPayload, 1000)

		manager := &actionManager{
			Name:        name,
			Action:      action,
			PayloadChan: ingestionMap[name],
		}
		go manager.start()
	}
}

func Execute(payloads []ActionPayload) {
	if len(payloads) == 0 {
		return
	}

	for _, payload := range payloads {
		if _, ok := ingestionMap[payload.Type]; ok {
			ingestionMap[payload.Type] <- payload
		}
	}
}

func (a *actionManager) start() {
	// TODO scale number of action managers as size of payload chan increases
	for action := range a.PayloadChan {
		a.Action.Execute(action)
	}
}
