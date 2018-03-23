package logger

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
)

const (
	ERROR   = 0
	WARNING = 1
	INFO    = 2
	VERBOSE = 3
)

type Logger struct {
	Name     string
	LogLevel int
}

func New(name string, logLevel int) Logger {
	return Logger{
		Name:     name,
		LogLevel: logLevel,
	}
}

func Log(message string, messageLevel int, l *Logger) {
	if l.LogLevel >= messageLevel {
		if len(l.Name) == 0 {
			log.Printf(message)
		} else {
			log.Printf("[%s] %s", l.Name, message)
		}
	}
}

func (l *Logger) Error(message string, a ...interface{}) {
	if a != nil {
		Log(color.HiRedString(message, a...), ERROR, l)
	} else {
		Log(color.HiRedString(message), ERROR, l)
	}
}

func (l *Logger) Warn(message string, a ...interface{}) {
	if a != nil {
		Log(color.HiYellowString(message, a...), WARNING, l)
	} else {
		Log(color.HiYellowString(message), WARNING, l)
	}
}

func (l *Logger) Info(message string, a ...interface{}) {
	if a != nil {
		Log(color.HiBlueString(message, a...), INFO, l)
	} else {
		Log(color.HiBlueString(message), INFO, l)
	}
}

func (l *Logger) Verbose(message string, a ...interface{}) {
	if a != nil {
		Log(fmt.Sprintf(message, a...), VERBOSE, l)
	} else {
		Log(fmt.Sprintf(message), VERBOSE, l)
	}
}

func (l *Logger) Fatal(message string, a ...interface{}) {
	l.Error(message, a)
	os.Exit(1)
}
