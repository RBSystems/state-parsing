package logger

var std = New("", VERBOSE)

func Error(message string, a ...interface{}) {
	std.Error(message, a...)
}

func Warn(message string, a ...interface{}) {
	std.Warn(message, a...)
}

func Info(message string, a ...interface{}) {
	std.Info(message, a...)
}

func Verbose(message string, a ...interface{}) {
	std.Verbose(message, a...)
}
