package dragon

import (
	"github.com/tauraamui/dragondaemon/internal/config"
	"github.com/tauraamui/dragondaemon/pkg/camera"
)

type Server interface {
	ConnectToCameras() []error
	LoadConfiguration()
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
	return camera.Connect(title, addr, sett)
}

func (s *server) LoadConfiguration() {
	s.config = config.Load()
}

func (s *server) Shutdown() error { return nil }
