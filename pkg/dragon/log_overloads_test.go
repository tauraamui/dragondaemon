package dragon_test

import "github.com/tauraamui/dragondaemon/pkg/log"

func overloadDebugLog(overload func(string, ...interface{})) func() {
	logDebugRef := log.Debug
	log.Debug = overload
	return func() { log.Info = logDebugRef }
}

func overloadWarnLog(overload func(string, ...interface{})) func() {
	logWarnRef := log.Warn
	log.Warn = overload
	return func() { log.Warn = logWarnRef }
}

func overloadInfoLog(overload func(string, ...interface{})) func() {
	logInfoRef := log.Info
	log.Info = overload
	return func() { log.Info = logInfoRef }
}
