package main

import (
	"flag"
	"fmt"

	"github.com/tacusci/logging"
	"gocv.io/x/gocv"
)

var shuttingDown bool

type options struct {
	debug         bool
	logFileName   string
	cameraAddress string
}

func parseCmdArgs() *options {
	opts := options{}

	flag.BoolVar(&opts.debug, "debug", false, "Set runtime mode to debug")
	flag.StringVar(&opts.logFileName, "log", "", "Server log file location")
	flag.StringVar(&opts.cameraAddress, "c", "", "RTSP address of camera to retrieve stream from")

	flag.Parse()

	loggingLevel := logging.WarnLevel
	logging.ColorLogLevelLabelOnly = true

	if opts.debug {
		logging.SetLevel(logging.DebugLevel)
		return &opts
	}

	logging.SetLevel(loggingLevel)

	return &opts
}

func main() {
	opts := parseCmdArgs()

	flushInitialised := make(chan bool)
	if len(opts.logFileName) > 0 {
		go logging.FlushLogs(opts.logFileName, &flushInitialised)
		//halt main thread until creating file to flush logs to has initialised
		<-flushInitialised
	}

	logging.WhiteOutput(fmt.Sprintf("Dragon Daemon v0.0.0\n"))

	camera, err := gocv.OpenVideoCapture(opts.cameraAddress)
	if err != nil {
		logging.ErrorAndExit(fmt.Sprintf("Connection to stream at [%s] has failed: %v\n", opts.cameraAddress, err))
	}
	defer camera.Close()

	logging.Info(fmt.Sprintf("Connected to stream at [%s]", opts.cameraAddress))
}
