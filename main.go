package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/tacusci/logging/v2"
	"github.com/takama/daemon"
	"github.com/tauraamui/dragondaemon/config"
	"github.com/tauraamui/dragondaemon/media"
)

const (
	name        = "dragon_daemon"
	description = "Dragon service daemon which saves RTSP media streams to disk"
)

type Service struct {
	daemon.Daemon
}

func (service *Service) Manage() (string, error) {
	usage := "Usage: dragond install | remove | start | stop | status"

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	logging.Info("Starting dragon daemon...")

	mediaServer := media.NewServer()
	cfg := config.Load()

	for _, c := range cfg.Cameras {
		if c.Disabled {
			logging.Warn(fmt.Sprintf("WARN: Connection %s is disabled, skipping...\n", c.Title))
			continue
		}

		mediaServer.Connect(
			c.Title,
			c.Address,
			c.PersistLoc,
			c.FPS,
			c.SecondsPerClip,
		)
	}

	mediaServer.BeginStreaming()
	wg := sync.WaitGroup{}
	go mediaServer.SaveStreams(&wg)

	killSignal := <-interrupt
	logging.Error(fmt.Sprintf("Received signal: %s", killSignal))

	mediaServer.Shutdown()
	logging.Warn("Waiting for persist process...")
	wg.Wait()
	logging.Info("Persist process has finished...")
	err := mediaServer.Close()
	if err != nil {
		logging.Error(fmt.Sprintf("Safe shutdown unsuccessful: %v\n", err))
		os.Exit(1)
	}

	return "Shutdown successful... BYE! 👋", nil
}

func init() {
	logging.ColorLogLevelLabelOnly = true
}

func main() {
	daemonType := daemon.SystemDaemon
	if runtime.GOOS == "darwin" {
		daemonType = daemon.UserAgent
	}

	srv, err := daemon.New(name, description, daemonType)
	if err != nil {
		logging.Error(err.Error())
		os.Exit(1)
	}

	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		logging.Error(fmt.Sprint(status, err.Error()))
		os.Exit(1)
	}

	fmt.Println(status)
}
