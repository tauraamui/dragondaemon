package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/tacusci/logging/v2"
	"github.com/takama/daemon"
	"github.com/tauraamui/dragondaemon/api"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/config"
	db "github.com/tauraamui/dragondaemon/pkg/database"
	"github.com/tauraamui/dragondaemon/pkg/dragon"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/media"
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
	logging.Info("Setting up dragondaemon service...") //nolint

	err := config.Setup()
	if err != nil {
		if !errors.Is(err, config.ErrConfigAlreadyExists) {
			return "", err
		}
		logging.Error(err.Error()) //nolint
	}

	err = db.Setup()
	if err != nil {
		if !errors.Is(err, db.ErrDBAlreadyExists) {
			return "", err
		}
		logging.Error(err.Error()) //nolint
	}

	return "Setup successful...", nil
}

func (service *Service) RemoveSetup() (string, error) {
	logging.Info("Removing setup for dragondaemon service...") //nolint
	err := db.Destroy()
	if err != nil {
		logging.Error("unable to delete database file: %s", err.Error()) //nolint
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

	logging.Info("Starting dragon daemon...") //nolint

	mediaServer := media.NewServer(debugMode)

	server := dragon.NewServer()
	err := server.LoadConfiguration()
	if err != nil {
		log.Fatal("unable to load config: %v", err)
	}
	errs := server.ConnectToCameras()
	for _, err := range errs {
		log.Error(err.Error())
	}

	cfg := config.New()
	logging.Info("Loading configuration") //nolint
	err = cfg.Load()
	if err != nil {
		logging.Fatal("Error loading configuration: %v", err) //nolint
	}
	logging.Info("Loaded configuration") //nolint

	for _, c := range cfg.Cameras {
		if c.Disabled {
			logging.Warn("Connection %s is disabled, skipping...", c.Title) //nolint
			continue
		}

		cameraConn, err := camera.Connect("Test", c.Address, camera.Settings{})
		if err != nil {
			logging.Fatal("error connecting to camera: %v", err)
		}
		cameraConn.Read()
		cameraConn.Close()

		settings := media.ConnectonSettings{
			PersistLocation: c.PersistLoc,
			MockWriter:      c.MockWriter,
			MockCapturer:    c.MockCapturer,
			FPS:             c.FPS,
			SecondsPerClip:  c.SecondsPerClip,
			DateTimeLabel:   c.DateTimeLabel,
			DateTimeFormat:  c.DateTimeFormat,
			Schedule:        c.Schedule,
			Reolink:         c.ReolinkAdvanced,
		}

		if err := settings.Validate(); err != nil {
			logging.Fatal(fmt.Errorf("settings validation failed: %w", err).Error())
		}

		mediaServer.Connect(
			c.Title,
			c.Address,
			settings,
		)
	}

	rpcListenPort := os.Getenv("DRAGON_RPC_PORT")
	if len(rpcListenPort) == 0 || !strings.Contains(rpcListenPort, ":") {
		rpcListenPort = ":3121"
	}

	logging.Info("Running API server on port %s...", rpcListenPort) //nolint
	mediaServerAPI, err := api.New(
		interrupt,
		mediaServer,
		api.Options{
			RPCListenPort: rpcListenPort,
			SigningSecret: cfg.Secret,
		},
	)
	if err != nil {
		logging.Error("unable to start API server: %v", err) //nolint
	} else {
		err := api.StartRPC(mediaServerAPI)
		if err != nil {
			logging.Error("Unable to start API RPC server: %v...", err) //nolint
		}
	}

	logging.Info("Running media server...") //nolint
	mediaServer.Run(media.Options{
		MaxClipAgeInDays: cfg.MaxClipAgeInDays,
	})

	// wait for application terminate signal from OS
	killSignal := <-interrupt
	fmt.Print("\r")
	logging.Error("Received signal: %s", killSignal) //nolint

	logging.Info("Shutting down API server...") //nolint
	err = api.ShutdownRPC(mediaServerAPI)
	if err != nil {
		logging.Error("Unable to shutdown API server: %v...", err) //nolint
	}

	// trigger server shutdown and wait
	logging.Info("Shutting down media server...") //nolint
	<-mediaServer.Shutdown()

	logging.Info("Closing camera connections...") //nolint
	err = mediaServer.Close()
	if err != nil {
		logging.Error(fmt.Sprintf("Safe shutdown unsuccessful: %v", err)) //nolint
		os.Exit(1)
	}

	return "Shutdown successful... BYE! 👋", nil
}

var debugMode bool

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
		debugMode = true
		logging.CurrentLoggingLevel = logging.DebugLevel
		logging.CallbackLabel = true
	default:
		logging.CurrentLoggingLevel = logging.InfoLevel
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
