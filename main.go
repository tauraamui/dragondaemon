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

var shuttingDown bool

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

	if opts.debug {
		logging.SetLevel(logging.DebugLevel)
		return &opts
	}

	logging.SetLevel(loggingLevel)

	if opts.secondsPerClip > 0 {
		opts.secondsPerClip++
	}

	return &opts
}

func main() {
	opts := parseCmdArgs()

	cfg := config.Load()
	for i, c := range cfg.Cameras {
		if c.Disabled == false {
			fmt.Printf("[%d] CAMERA: %v\n", i, c.Title)
		}
	}

	flushInitialised := make(chan bool)
	if len(opts.logFileName) > 0 {
		go logging.FlushLogs(opts.logFileName, &flushInitialised)
		//halt main thread until creating file to flush logs to has initialised
		<-flushInitialised
	}

	logging.WhiteOutput(fmt.Sprintf("Dragon Daemon v0.0.0\n"))

	mediaServer := media.NewServer()
	go listenForStopSig(mediaServer)

	for mediaServer.IsRunning() {
	}

	// go func() {
	// 	conn, err := mediaServer.Connect("Front", opts.cameraAddress)
	// 	if err == nil {
	// 		for mediaServer.IsRunning() {
	// 			conn.PersistToDisk(opts.persistLocationPath, opts.secondsPerClip)
	// 		}
	// 	}
	// }()

	// conn, err := mediaServer.Connect("BigBuckBunny", "rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov")
	// if err == nil {
	// 	for mediaServer.IsRunning() {
	// 		conn.PersistToDisk(opts.persistLocationPath, opts.secondsPerClip)
	// 	}
	// }
}

func listenForStopSig(srv *media.Server) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	sig := <-gracefulStop
	// logging.Debug("Stopping, clearing old sessions...")
	//send a terminate command to the session clearing goroutine's channel
	shuttingDown = true
	logging.Error(fmt.Sprintf("☠️ Caught sig: %+v (Shutting down and cleaning up...) ☠️", sig))
	logging.Info("Stopping media server...")
	srv.Shutdown()
}
