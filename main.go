package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/takama/daemon"
	"github.com/tauraamui/dragondaemon/config"
	"github.com/tauraamui/dragondaemon/media"
)

const (
	name        = "dragon_daemon"
	description = "Dragon daemon saves RTSP media streams to disk"
)

var stdlog, errlog *log.Logger

type Service struct {
	daemon.Daemon
}

func (service *Service) Manage() (string, error) {
	usage := "Usage: dragond install | remove | start | stop | status"

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			service.Install()
		case "remove":
			service.Remove()
		case "start":
			service.Start()
		case "stop":
			service.Stop()
		case "status":
			service.Status()
		default:
			return usage, nil
		}
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	mediaServer := media.NewServer()
	cfg := config.Load(stdlog, errlog)

	for _, c := range cfg.Cameras {
		if c.Disabled {
			stdlog.Printf("WARN: Connection %s is disabled, skipping...\n", c.Title)
			continue
		}

		mediaServer.Connect(
			stdlog, errlog,
			c.Title,
			c.Address,
			c.PersistLoc,
			c.SecondsPerClip,
		)
	}

	mediaServer.BeginStreaming()
	wg := sync.WaitGroup{}
	go mediaServer.SaveStreams(&wg)

	killSignal := <-interrupt
	stdlog.Println("Received signal:", killSignal)

	stdlog.Println("Waiting for persist process...")
	wg.Wait()
	stdlog.Println("Persist process has finished...")
	mediaServer.Shutdown()
	err := mediaServer.Close()
	if err != nil {
		errlog.Printf("Safe shutdown unsuccessful: %v\n", err)
		os.Exit(1)
	}

	return "Shutdown successful... BYE! �", nil
}

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

func main() {
	daemonType := daemon.SystemDaemon
	if runtime.GOOS == "darwin" {
		daemonType = daemon.GlobalDaemon
	}

	srv, err := daemon.New(name, description, daemonType)
	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}

	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		errlog.Println(status, "\nError ", err)
		os.Exit(1)
	}

	fmt.Println(status)
}
