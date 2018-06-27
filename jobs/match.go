package jobs

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
)

var EventMatchStream chan elkreporting.ElkEvent

type MatchConfig struct {
	Hostname         string `json:"hostname,omitempty"`
	Timestamp        string `json:"timestamp,omitempty"`
	LocalEnvironment string `json:"localEnvironment,omitempty"`
	Building         string `json:"building,omitempty"`
	Room             string `json:"room,omitempty"`

	Event struct {
		Type           string `json:"type,omitempty"`
		Requestor      string `json:"requestor,omitempty"`
		EventCause     string `json:"eventCause,omitempty"`
		Device         string `json:"device,omitempty"`
		EventInfoKey   string `json:"eventInfoKey,omitempty"`
		EventInfoValue string `json:"eventInfoValue,omitempty"`
	} `json:"event,omitempty"`

	Regex struct {
		Hostname         *regexp.Regexp
		Timestamp        *regexp.Regexp
		LocalEnvironment *regexp.Regexp
		Building         *regexp.Regexp
		Room             *regexp.Regexp

		Event struct {
			Type           *regexp.Regexp
			Requestor      *regexp.Regexp
			EventCause     *regexp.Regexp
			Device         *regexp.Regexp
			EventInfoKey   *regexp.Regexp
			EventInfoValue *regexp.Regexp
		}
	}
}

var matchJobsMux sync.RWMutex
var matchJobs []*runner

func startEventMatcher() {
	log.L.Infof("started event matcher")
	EventMatchStream = make(chan elkreporting.ElkEvent, 1000)

	for event := range EventMatchStream {
		log.L.Debugf("[event-matcher] received event: %+v", event)

		matchJobsMux.RLock()
		for _, runner := range matchJobs {
			if shouldRunJob(runner, event) {
				log.L.Infof("[%s|%v] Running job from event...", runner.Config.Name, runner.TriggerIndex)
				go runner.runJob()
			}
		}
		matchJobsMux.RUnlock()
	}
}

func addMatchJob(job *runner) {
	matchJobsMux.Lock()
	matchJobs = append(matchJobs, job)
	matchJobsMux.Unlock()
}

func shouldRunJob(job *runner, event elkreporting.ElkEvent) bool {
	if job.Trigger.Match.Regex.Hostname != nil {
		passed := job.Trigger.Match.Regex.Hostname.MatchString(event.Hostname)
		if !passed {
			return false
		}
	}

	if job.Trigger.Match.Regex.Timestamp != nil {
		passed := job.Trigger.Match.Regex.Timestamp.MatchString(event.Timestamp)
		if !passed {
			return false
		}
	}

	if job.Trigger.Match.Regex.LocalEnvironment != nil {
		passed := job.Trigger.Match.Regex.LocalEnvironment.MatchString(fmt.Sprintf("%v", event.LocalEnvironment))
		if !passed {
			return false
		}
	}

	if job.Trigger.Match.Regex.Building != nil {
		passed := job.Trigger.Match.Regex.Building.MatchString(event.Building)
		if !passed {
			return false
		}
	}

	if job.Trigger.Match.Regex.Room != nil {
		passed := job.Trigger.Match.Regex.Room.MatchString(event.Room)
		if !passed {
			return false
		}
	}

	if job.Trigger.Match.Regex.Event.Type != nil {
		passed := job.Trigger.Match.Regex.Event.Type.MatchString(fmt.Sprintf("%v", event.Event.Event.Type))
		if !passed {
			return false
		}
	}

	if job.Trigger.Match.Regex.Event.Requestor != nil {
		passed := job.Trigger.Match.Regex.Event.Requestor.MatchString(event.Event.Event.Requestor)
		if !passed {
			return false
		}
	}

	if job.Trigger.Match.Regex.Event.EventCause != nil {
		passed := job.Trigger.Match.Regex.Event.EventCause.MatchString(fmt.Sprintf("%v", event.Event.Event.EventCause))
		if !passed {
			return false
		}
	}

	if job.Trigger.Match.Regex.Event.Device != nil {
		passed := job.Trigger.Match.Regex.Event.Device.MatchString(event.Event.Event.Device)
		if !passed {
			return false
		}
	}

	if job.Trigger.Match.Regex.Event.EventInfoKey != nil {
		passed := job.Trigger.Match.Regex.Event.EventInfoKey.MatchString(event.Event.Event.EventInfoKey)
		if !passed {
			return false
		}
	}

	if job.Trigger.Match.Regex.Event.EventInfoValue != nil {
		passed := job.Trigger.Match.Regex.Event.EventInfoValue.MatchString(event.Event.Event.EventInfoValue)
		if !passed {
			return false
		}
	}

	return true
}
