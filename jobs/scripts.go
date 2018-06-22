package jobs

import (
	"os"
	"os/exec"

	"github.com/byuoitav/common/log"
)

type ScriptJob struct {
	Path string
}

func (j *ScriptJob) Run() []Action {
	if len(j.Path) == 0 {
		log.L.Errorf("path for a script job wasn't set. can't run this job.")
		return []Action{}
	}

	// build the cmd
	cmd := exec.Command(j.Path)
	cmd.Stdout = os.Stdout // TODO these should match wherever the logger is going
	cmd.Stderr = os.Stderr

	// execute script
	log.L.Infof("Executing script %s", j.Path)
	err := cmd.Run()
	if err != nil {
		log.L.Warnf("error executing script %s: %s", j.Path, err)
	}

	log.L.Infof("Script %s ran successfuly.")
	return []Action{}
}
