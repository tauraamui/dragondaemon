package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tacusci/logging"
	"github.com/tauraamui/dragondaemon/config"
	"github.com/tauraamui/dragondaemon/media"
)

type options struct {
	debug               bool
	logFileName         string
	cameraAddress       string
	secondsPerClip      uint
	persistLocationPath string
}

func parseCmdArgs() *options {
	opts := options{}

	flag.BoolVar(&opts.debug, "debug", false, "Set runtime mode to debug")
	flag.StringVar(&opts.logFileName, "log", "", "Server log file location")
	flag.StringVar(&opts.cameraAddress, "c", "", "RTSP address of camera to retrieve stream from")
	flag.UintVar(&opts.secondsPerClip, "s", 30, "Number of seconds per video clip")
	flag.StringVar(&opts.persistLocationPath, "d", "", "Directory to store video clips")

	flag.Parse()

	loggingLevel := logging.WarnLevel
	logging.OutputPath = false
	logging.ColorLogLevelLabelOnly = true

	logging.SetLevel(loggingLevel)

	if opts.debug {
		logging.SetLevel(logging.DebugLevel)
		return &opts
	}

	if opts.secondsPerClip > 0 {
		opts.secondsPerClip++
	}

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

	logging.WhiteOutput(fmt.Sprintf("Starting Dragon Daemon v0.0.0 (c)[tacusci ltd]\n"))

	mediaServer := media.NewServer()

	go listenForStopSig(mediaServer)

	cfg := config.Load()

	for _, c := range cfg.Cameras {
		if c.Disabled {
			logging.Warn(fmt.Sprintf("Connection %s is disabled, skipping...", c.Title))
			continue
		}

		mediaServer.Connect(
			c.Title,
			c.Address,
			c.PersistLoc,
			c.SecondsPerClip,
		)
	}

	mediaServer.BeginStreaming()
	mediaServer.SaveStreams()

	err := mediaServer.Close()
	if err != nil {
		logging.Error(fmt.Sprintf("Safe shutdown unsuccessful: %v", err))
		os.Exit(1)
	}
	logging.Info("Shutdown successful... BYE! 👋")
}

func listenForStopSig(srv *media.Server) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	sig := <-gracefulStop
	//send a terminate command to the session clearing goroutine's channel
	logging.Error(fmt.Sprintf("☠️ Caught sig: %+v (Shutting down and cleaning up...) ☠️", sig))
	logging.Info("Stopping media server...")
	srv.Shutdown()
	logging.Info("Closing stream connections...")
}
