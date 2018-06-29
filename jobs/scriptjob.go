package jobs

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/state-parsing/actions"
)

type ScriptJob struct {
	Path string
}

func (j *ScriptJob) Run(ctx interface{}) []actions.ActionPayload {
	if len(j.Path) == 0 {
		log.L.Errorf("path for a script job wasn't set. can't run this job.")
		return []actions.ActionPayload{}
	}

	// add context for timeout
	contxt, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// build the cmd
	cmd := exec.CommandContext(contxt, j.Path)
	cmd.Stdout = os.Stdout // TODO these should match wherever the logger is going
	cmd.Stderr = os.Stderr

	// execute script
	log.L.Infof("Executing script %s", j.Path)
	err := cmd.Run()
	if err != nil {
		log.L.Warnf("error executing script %s: %s", j.Path, err)
		return []actions.ActionPayload{}
	}

	log.L.Infof("Script %s ran successfuly.")
	return []actions.ActionPayload{}
}
