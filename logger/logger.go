package logger

import (
	"fmt"
	"log"

	"github.com/fatih/color"
)

const (
	ERROR   = 0
	WARNING = 1
	INFO    = 2
	VERBOSE = 3
)

type Logger struct {
	LogLevel int
	Name     string
}

// logs any message to a specific messageLevel
func L(message string, messageLevel int, l *Logger) {
	if l.LogLevel >= messageLevel {
		if len(l.Name) == 0 {
			log.Printf(message)
		} else {
			log.Printf("[%s] %s", l.Name, message)
		}
	}
}

// Error. These will always be logged.
func (l *Logger) E(message string, a ...interface{}) {
	if a != nil {
		L(color.HiRedString(message, a...), ERROR, l)
	} else {
		L(color.HiRedString(message), ERROR, l)
	}
}

func (l *Logger) Error(message string, a ...interface{}) {
	l.E(message, a...)
}

// Warn
func (l *Logger) W(message string, a ...interface{}) {
	if a != nil {
		L(color.HiYellowString(message, a...), WARNING, l)
	} else {
		L(color.HiYellowString(message), WARNING, l)
	}
}

func (l *Logger) Warn(message string, a ...interface{}) {
	l.W(message, a...)
}

// Info
func (l *Logger) I(message string, a ...interface{}) {
	if a != nil {
		L(color.HiBlueString(message, a...), INFO, l)
	} else {
		L(color.HiBlueString(message), INFO, l)
	}
}
func (l *Logger) Info(message string, a ...interface{}) {
	l.I(message, a...)
}

// Verbose
func (l *Logger) V(message string, a ...interface{}) {
	if a != nil {
		L(fmt.Sprintf(message, a...), VERBOSE, l)
	} else {
		L(fmt.Sprintf(message), VERBOSE, l)
	}
}

func (l *Logger) Verbose(message string, a ...interface{}) {
	l.V(message, a...)
}
