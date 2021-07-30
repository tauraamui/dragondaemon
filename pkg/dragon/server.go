package dragon

import (
	"github.com/tauraamui/dragondaemon/internal/config"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/log"
)

type Server interface {
	ConnectToCameras() []error
	LoadConfiguration() error
	Shutdown() error
}

func NewServer() Server {
	return &server{}
}

type server struct {
	cameras []camera.Connection
	config  config.Values
}

func (s *server) ConnectToCameras() []error {
	var errs []error

	for _, cam := range s.config.Cameras {
		if cam.Disabled {
			log.Warn("Camera [%s] is disabled... skipping...", cam.Title)
			continue
		}
		conn, err := connectToCamera(cam.Title, cam.Address, camera.Settings{})
		if err != nil {
			errs = append(errs, err)
		}

		if conn != nil {
			s.cameras = append(s.cameras, conn)
		}
	}
	return errs
}

func connectToCamera(title, addr string, sett camera.Settings) (camera.Connection, error) {
	log.Debug("connecting to camera: %s", title)
	return camera.Connect(title, addr, sett)
}

func (s *server) LoadConfiguration() error {
	config, err := config.Load()
	if err != nil {
		return err
	}

	s.config = config
	return nil
}

func (s *server) Shutdown() error { return nil }
