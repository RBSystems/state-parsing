package updater

import (
	"log"

	"github.com/fatih/color"
)

/*
 * If you implement this interface, you will be able to run as an updater.
 */
type Interface interface {
	Run(loggingLevel int) error
}

/*
 * If you add embed this struct, your `run()` function will be run after some default logging/logic built into all Updaters, as well as logic after execution.
 */
type Updater struct {
	mu ManagedUpdater

	Name     string
	LogLevel int
}

type ManagedUpdater interface {
	Init()
	Run() error
}

func NewUpdater(mu ManagedUpdater) *Updater {
	u := &Updater{
		mu: mu,
	}

	u.mu.Init()

	return u
}

func (u *Updater) Run(loggingLevel int) error {
	if u.LogLevel != loggingLevel {
		log.Printf(color.HiYellowString("[%s] Changing logging level to %v", u.Name, loggingLevel))
		u.LogLevel = loggingLevel
	}

	return u.mu.Run()
}
