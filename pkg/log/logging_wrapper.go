package log

import "github.com/tacusci/logging/v2"

var Debug = func(format string, a ...interface{}) {
	logging.Debug(format, a...) //nolint
}

var Info = func(format string, a ...interface{}) {
	logging.Info(format, a...) //nolint
}

var Warn = func(format string, a ...interface{}) {
	logging.Warn(format, a...) //nolint
}

var Error = func(format string, a ...interface{}) {
	logging.Error(format, a...) //nolint
}

var Fatal = func(format string, a ...interface{}) {
	logging.Fatal(format, a...) //nolint
}
