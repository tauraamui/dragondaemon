package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/tacusci/logging/v2"
	"github.com/takama/daemon"
	"github.com/tauraamui/dragondaemon/api"
	"github.com/tauraamui/dragondaemon/config"
	db "github.com/tauraamui/dragondaemon/data"
	"github.com/tauraamui/dragondaemon/media"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	name        = "dragon_daemon"
	description = "Dragon service daemon which saves RTSP media streams to disk"
	success     = "\t\t\t\t\t[  \033[32mOK\033[0m  ]" // Show colored "OK"
	failed      = "\t\t\t\t\t[\033[31mFAILED\033[0m]" // Show colored "FAILED"
)

type Service struct {
	daemon.Daemon
}

// Setup will setup local DB and ask for root admin credentials
func (service *Service) Setup() (string, error) {
	logging.Info("Setting up dragondaemon service...")
	err := db.Create()
	if err != nil {
		return "", err
	}

	stdinReader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter username: ")
	username, _ := stdinReader.ReadString('\n')

	fmt.Printf("Enter password: ")
	passwordBytes, err := terminal.ReadPassword(0)
	if err != nil {
		return "", err
	}

	fmt.Println()
	err = db.CreateRootUser(username, string(passwordBytes))
	if err != nil {
		return "", err
	}
	return "Setup successful...", nil
}

func (service *Service) RemoveSetup() (string, error) {
	logging.Info("Removing setup for dragondaemon service...")
	err := db.Destroy()
	if err != nil {
		return "", err
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

	logging.Info("Starting dragon daemon...")

	mediaServer := media.NewServer(debugMode)

	cfg := config.New()
	logging.Info("Loading configuration")
	err := cfg.Load()
	if err != nil {
		logging.Fatal("Error loading configuration: %v", err)
	}
	logging.Info("Loaded configuration")

	for _, c := range cfg.Cameras {
		if c.Disabled {
			logging.Warn("Connection %s is disabled, skipping...", c.Title)
			continue
		}

		mediaServer.Connect(
			c.Title,
			c.Address,
			c.PersistLoc,
			c.FPS,
			c.SecondsPerClip,
			c.Schedule,
			c.ReolinkAdvanced,
		)
	}

	rpcListenPort := os.Getenv("DRAGON_RPC_PORT")
	if len(rpcListenPort) == 0 || !strings.Contains(rpcListenPort, ":") {
		rpcListenPort = ":3121"
	}
	logging.Info("Running API server on port %s...", rpcListenPort)
	mediaServerAPI := api.New(
		interrupt,
		mediaServer,
		api.Options{RPCListenPort: rpcListenPort},
	)
	err = api.StartRPC(mediaServerAPI)
	if err != nil {
		logging.Error("Unable to start API RPC server: %v...", err)
	}

	logging.Info("Running media server...")
	mediaServer.Run(media.Options{
		MaxClipAgeInDays: cfg.MaxClipAgeInDays,
	})

	// wait for application terminate signal from OS
	killSignal := <-interrupt
	fmt.Print("\r")
	logging.Error("Received signal: %s", killSignal)

	logging.Info("Shutting down API server...")
	err = api.ShutdownRPC(mediaServerAPI)
	if err != nil {
		logging.Error("Unable to shutdown API server: %v...", err)
	}

	// trigger server shutdown and wait
	logging.Info("Shutting down media server...")
	<-mediaServer.Shutdown()

	logging.Info("Closing camera connections...")
	err = mediaServer.Close()
	if err != nil {
		logging.Error(fmt.Sprintf("Safe shutdown unsuccessful: %v", err))
		os.Exit(1)
	}

	return "Shutdown successful... BYE! 👋", nil
}

var debugMode bool

func init() {
	logging.CallbackLabelLevel = 4
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
		logging.Error(err.Error())
		os.Exit(1)
	}

	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		logging.Error(err.Error())
		os.Exit(1)
	}

	logging.Info(status)
}
