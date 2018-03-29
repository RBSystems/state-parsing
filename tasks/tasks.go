package tasks

import (
	"github.com/byuoitav/state-parsing/alerts/heartbeat"
	"github.com/byuoitav/state-parsing/tasks/names"
	"github.com/byuoitav/state-parsing/tasks/task"
	"github.com/byuoitav/state-parsing/update/roomupdate"
)

var tasks map[string]task.Task

func init() {
	tasks = make(map[string]task.Task)

	// add all jobs here
	tasks[names.ROOM_UPDATE] = task.NewTask(&roomupdate.RoomUpdater{})
	tasks[names.LOST_HEARTBEAT] = task.NewTask(&heartbeat.LostHeartbeatAlertFactory{})
	tasks[names.HEARTBEAT_RESTORED] = task.NewTask(&heartbeat.RestoredHeartbeatAlertFactory{})
}

func GetTask(name string) (task.Task, bool) {
	task, ok := tasks[name]
	return task, ok
}
