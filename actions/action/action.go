package action

import "github.com/byuoitav/common/nerr"

type Action struct {
	Type    string // type of the alert, found in constants above
	Device  string // the device the alert corresponds to
	Content interface{}
}

type Result struct {
	Action
	Error *nerr.E
}
