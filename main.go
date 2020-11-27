package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tacusci/logging"
	"gocv.io/x/gocv"
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

	flushInitialised := make(chan bool)
	if len(opts.logFileName) > 0 {
		go logging.FlushLogs(opts.logFileName, &flushInitialised)
		//halt main thread until creating file to flush logs to has initialised
		<-flushInitialised
	}

	logging.WhiteOutput(fmt.Sprintf("Dragon Daemon v0.0.0"))

	camera, err := gocv.OpenVideoCapture(opts.cameraAddress)
	if err != nil {
		logging.ErrorAndExit(fmt.Sprintf("Connection to stream at [%s] has failed: %v", opts.cameraAddress, err))
	}
	defer camera.Close()

	logging.Info(fmt.Sprintf("Connected to stream at [%s]", opts.cameraAddress))

	img := gocv.NewMat()
	defer img.Close()

	if ok := camera.Read(&img); !ok {
		logging.ErrorAndExit(fmt.Sprintf("Unable to read from stream at [%s]\n", opts.cameraAddress))
	}

	outputFile := fetchClipFilePath(opts.persistLocationPath)
	writer, err := gocv.VideoWriterFile(outputFile, "MJPG", 30, img.Cols(), img.Rows(), true)
	if err != nil {
		logging.Error(fmt.Sprintf("Opening video writer device: %v\n", err))
	}
	defer writer.Close()

	var framesWritten uint
	for framesWritten = 0; framesWritten < 30*opts.secondsPerClip; framesWritten++ {
		if ok := camera.Read(&img); !ok {
			logging.Error(fmt.Sprintf("Device for stream at [%s] closed", opts.cameraAddress))
			return
		}
		if img.Empty() {
			logging.Debug("Skipping frame...")
			continue
		}

		if err := writer.Write(img); err != nil {
			logging.Error(fmt.Sprintf("Unable to write frame to file: %v", err))
		}
	}
}

func fetchClipFilePath(dirPath string) string {
	if len(dirPath) > 0 {
		ensureDirectoryExists(dirPath)
	} else {
		dirPath = "."
	}

	return filepath.FromSlash(fmt.Sprintf("%s/%s.avi", dirPath, time.Now().Format("2006-01-02 15.04.05")))
}

func ensureDirectoryExists(path string) error {
	err := os.Mkdir(path, os.ModePerm)

	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}
