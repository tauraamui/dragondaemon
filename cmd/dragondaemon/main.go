package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/tacusci/logging/v2"
	"github.com/takama/daemon"
	"github.com/tauraamui/dragondaemon/pkg/config"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	db "github.com/tauraamui/dragondaemon/pkg/database"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
	"gocv.io/x/gocv"
)

const (
	name        = "dragon_daemon"
	description = "Dragon service daemon which saves RTSP media streams to disk"
)

type Service struct {
	daemon.Daemon
}

// Setup will setup local DB and ask for root admin credentials
func (service *Service) Setup() (string, error) {
	log.Info("Setting up dragondaemon service...")

	err := config.DefaultCreator().Create()
	if err != nil {
		if !errors.Is(err, configdef.ErrConfigAlreadyExists) {
			return "", err
		}
		log.Error(err.Error())
	}

	err = db.Setup()
	if err != nil {
		if !errors.Is(err, db.ErrDBAlreadyExists) {
			return "", err
		}
		log.Error(err.Error())
	}

	return "Setup successful...", nil
}

func (service *Service) RemoveSetup() (string, error) {
	log.Info("Removing setup for dragondaemon service...")
	err := db.Destroy()
	if err != nil {
		log.Error("unable to delete database file: %s", err.Error())
	}

	return "Removing setup successful...", nil
}

func (service *Service) Manage() (string, error) {
	usage := "Usage: dragond setup | remove-setup | install | remove | start | stop | status"

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "setup":
			return service.Setup()
		case "remove-setup":
			return service.RemoveSetup()
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

	log.Info("Starting dragon daemon...")

	server, err := dragon.NewServer(config.DefaultResolver(), video.ResolveBackend(os.Getenv("DRAGON_VIDEO_BACKEND")))
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx, cancelStartup := context.WithCancel(context.Background())
	go startupServer(ctx, server)

	killSignal := <-interrupt
	fmt.Print("\r")
	log.Error("Received signal: %s", killSignal)

	cancelStartup()
	log.Info("Shutting down server...")
	<-server.Shutdown()

	var b bytes.Buffer
	gocv.MatProfile.Count()
	gocv.MatProfile.WriteTo(&b, 1)
	fmt.Print(b.String())

	return "Shutdown successful... BYE! 👋", nil
}

func startupServer(ctx context.Context, server *dragon.Server) {
	connectToCameras(ctx, server)
	server.SetupProcesses()
	server.RunProcesses()
}

func connectToCameras(ctx context.Context, server *dragon.Server) {
	errs := server.ConnectWithCancel(ctx)
	for _, err := range errs {
		log.Error(err.Error())
	}
}

func init() {
	logging.CallbackLabelLevel = 5
	logging.ColorLogLevelLabelOnly = true
	loggingLevel := os.Getenv("DRAGON_LOGGING_LEVEL")

	switch strings.ToLower(loggingLevel) {
	case "info":
		logging.CurrentLoggingLevel = logging.InfoLevel
	case "warn":
		logging.CurrentLoggingLevel = logging.WarnLevel
	case "debug":
		logging.CurrentLoggingLevel = logging.DebugLevel
		logging.CallbackLabel = true
	default:
		logging.CurrentLoggingLevel = logging.WarnLevel
	}
}

func main() {
	daemonType := daemon.SystemDaemon
	if runtime.GOOS == "darwin" {
		daemonType = daemon.UserAgent
	}

	srv, err := daemon.New(name, description, daemonType)
	if err != nil {
		logging.Error(err.Error()) //nolint
		os.Exit(1)
	}

	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		logging.Error(err.Error()) //nolint
		os.Exit(1)
	}

	logging.Info(status) //nolint
}
