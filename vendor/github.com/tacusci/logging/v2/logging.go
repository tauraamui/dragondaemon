package logging

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/fatih/color"
)

type level struct {
	index int
	w     io.Writer
	p     *color.Color
	s     string
}

var (
	SilentLevel level = level{index: 0, w: nil, s: "", p: nil}
	DebugLevel  level = level{index: 1, w: color.Output, s: "DEBUG", p: color.New(color.FgYellow)}
	WarnLevel   level = level{index: 2, w: color.Output, s: "WARN", p: color.New(color.FgYellow)}
	InfoLevel   level = level{index: 3, w: color.Output, s: "INFO", p: color.New(color.FgGreen)}
	errorLevel  level = level{index: 4, w: color.Error, s: "ERROR", p: color.New(color.FgRed)}
)

var CurrentLoggingLevel level = InfoLevel
var ColorLogLevelLabelOnly = true
var CallbackLabel = false
var CallbackLabelLevel = 2

//SetLevel allows settings of the level of logging
func SetLevel(loggingLevel level) {
	CurrentLoggingLevel = loggingLevel
}

func log(logLevel level, format string, a ...interface{}) (n int, err error) {
	if CurrentLoggingLevel.index == 0 || CurrentLoggingLevel.index > logLevel.index {
		return
	}

	printFunc := logLevel.p.Fprintf
	strPrintFunc := logLevel.p.Sprintf
	colorInjector := logLevel.p.SprintFunc()
	if ColorLogLevelLabelOnly {
		printFunc = fmt.Fprintf
		strPrintFunc = fmt.Sprintf
	}

	return printFunc(
		logLevel.w,
		"%s [%s]%s %s\n",
		getTimeString(),
		colorInjector(logLevel.s),
		func() string {
			if CallbackLabel {
				return fmt.Sprintf("(%s)", createCallbackLabel(CallbackLabelLevel))
			}
			return ""
		}(),
		strPrintFunc(format, a...),
	)
}

//Info outputs log line to console with green color text
func Info(format string, a ...interface{}) (n int, err error) {
	return log(InfoLevel, format, a...)
}

//Warn outputs log line to console with yellow color text
func Warn(format string, a ...interface{}) (n int, err error) {
	return log(WarnLevel, format, a...)
}

//Debug outputs log line to console with yellow color text
func Debug(format string, a ...interface{}) (n int, err error) {
	return log(DebugLevel, format, a...)
}

//Error outputs log line to console with red color text
func Error(format string, a ...interface{}) (n int, err error) {
	return log(errorLevel, format, a...)
}

//ErrorAndExit outputs log line to console with red color text and exits
func Fatal(format string, a ...interface{}) {
	Error(format, a...)
	os.Exit(1)
}

func createCallbackLabel(skip int) string {
	function, _, line, _ := runtime.Caller(skip)
	return fmt.Sprintf("(%s):%d", runtime.FuncForPC(function).Name(), line)
}

//getTimeString gets formatted string to timestamp log and console output
func getTimeString() string {
	t := time.Now()
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}
