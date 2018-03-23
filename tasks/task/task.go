package task

import (
	"github.com/byuoitav/state-parsing/logger"
)

type Interface interface {
	/*
	 * Init will be called once for your struct during the life of the program
	 */
	Init()

	/*
	 * Pre will be called each time a Task is run, just before Run() is called.
	 *
	 * If you want to continue, return with a boolean of true.
	 * If you want the Task to stop, return false (likely if there is some error that will make something break)
	 */
	Pre() (error, bool)

	/*
	 * The main part of a task. Called each time the Task is run, unless Pre() returns false.
	 */
	Run() error

	/*
	 * Post is called after Run(), each time a Task is run, unless Pre() returns false.
	 * Gets passed the same error that returned from Run().
	 */
	Post(error)
}

/*
 * Compose this struct into another struct to make it a Task.
 */

type Task struct {
	i Interface
	logger.Logger
}

func NewTask(i Interface) Task {
	t := Task{
		i: i,
	}

	t.i.Init()

	return t
}

func (t *Task) Run(loggingLevel int) error {
	if t.LogLevel != loggingLevel {
		t.Error("Changing logging level to %v", loggingLevel)
		t.LogLevel = loggingLevel
	}

	err, cont := t.i.Pre()
	if err != nil || !cont {
		t.Error("error in PreRun: %s", err)

		if !cont {
			t.Error("quitting Run.")
			return nil
		}

		t.Warn("Continuing anyways...")
	}

	err = t.i.Run()
	if err != nil {
		t.Error("error in Run: %s", err)
	}

	t.i.Post(err)

	return err
}

/* Default methods for Tasks, in case you're lazy */

/*
 * This one should probably be overridden.
 */
func (t *Task) Init() {
	t.Logger = logger.New("Task", logger.VERBOSE)
}

func (t *Task) Pre() (error, bool) {
	return nil, true
}

func (t *Task) Post(err error) {
}
