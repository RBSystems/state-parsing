package actiongen

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/state-parser/actions/action"
	reflections "gopkg.in/oleiade/reflections.v1"
)

const (
	//Email .
	Email = "email"

	//Slack .
	Slack = "slack"
)

var paramRe *regexp.Regexp

func init() {
	paramRe = regexp.MustCompile(`{{.*?}}`)
}

var actionGenerators = map[string]func(Config, events.Event, string) (action.Payload, *nerr.E){
	Email: GenEmailAction,
	Slack: GenSlackAction,
}

//Config .
type Config struct {
	//Type is the type of action to be generated.
	Type string `json:"action-type"`

	//Fields is the mapping of how to take the fields of the event and feed them into the action generated. The expected keys are unique to the type of action.
	Fields map[string]string `json:"fields"`
}

//GenerateAction .
func GenerateAction(t Config, e events.Event, d string) (action.Payload, *nerr.E) {
	if v, ok := actionGenerators[t.Type]; ok {
		return v(t, e, d)
	}
	return action.Payload{}, nerr.Create(fmt.Sprintf("unkown type %v", t.Type), "invalid-type")
}

//ReplaceParameters .
func ReplaceParameters(s string, e events.Event) (string, *nerr.E) {

	//find if there are any parameters in s
	matches := paramRe.FindAllString(s, -1)
	if matches == nil {
		//no changes
		return s, nil
	}
	var inter interface{}
	var err error

	for i := range matches {
		inter = e
		cur := strings.Trim(matches[i], "{}")

		log.L.Debugf("getting field: %v", cur)
		//check to see if it's a sub field
		if strings.Contains(cur, ".") {
			split := strings.Split(cur, ".")
			for j := range split {
				inter, err = reflections.GetField(inter, split[j])
				if err != nil {
					return s, nerr.Translate(err).Addf("Unkown field for parameterization: %v", cur)
				}
			}
		} else {
			inter, err = reflections.GetField(inter, cur)
			if err != nil {
				return s, nerr.Translate(err).Addf("Unkown field for parameterization: %v", cur)
			}
		}
		log.L.Debugf("Got %v", inter)

		toReplace := fmt.Sprintf("%v", inter)
		s = strings.Replace(s, matches[i], toReplace, -1)
	}

	return s, nil
}
