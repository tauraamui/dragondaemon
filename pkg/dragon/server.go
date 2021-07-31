package dragon

import (
	"context"
	"sync"

	"github.com/tauraamui/dragondaemon/internal/config"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/log"
)

type Server interface {
	Connect() []error
	ConnectWithCancel(context.Context) []error
	LoadConfiguration() error
	Shutdown() chan interface{}
}

func NewServer() Server {
	return &server{}
}

type server struct {
	shutdownDone chan interface{}
	config       config.Values
	mu           sync.Mutex
	cameras      []camera.Connection
}

func (s *server) Connect() []error {
	return s.connect(context.Background())
}

func (s *server) ConnectWithCancel(cancel context.Context) []error {
	return s.connect(cancel)
}

type connectResult struct {
	cam camera.Connection
	err error
}

func (s *server) connect(cancel context.Context) []error {
	s.shutdownDone = make(chan interface{})
	connAndError := make(chan connectResult)
	var errs []error

	s.mu.Lock()
	defer s.mu.Unlock()

	wg := sync.WaitGroup{}
	wg.Add(len(s.config.Cameras))
	for _, cam := range s.config.Cameras {
		go func(cancel context.Context, wg *sync.WaitGroup, cam config.Camera, connAndError chan connectResult) {
			defer wg.Done()
			select {
			case <-cancel.Done():
				return
			default:
				if cam.Disabled {
					log.Warn("Camera [%s] is disabled... skipping...", cam.Title)
					return
				}
				settings := camera.Settings{
					DateTimeFormat:  cam.DateTimeFormat,
					DateTimeLabel:   cam.DateTimeLabel,
					FPS:             cam.FPS,
					PersistLocation: cam.PersistLoc,
					Reolink:         cam.ReolinkAdvanced,
				}
				conn, err := connectToCamera(cancel, cam.Title, cam.Address, settings)
				connAndError <- connectResult{
					cam: conn,
					err: err,
				}
			}
		}(cancel, &wg, cam, connAndError)
	}

	go func(wg *sync.WaitGroup, c chan connectResult) {
		wg.Wait()
		close(c)
	}(&wg, connAndError)

	for r := range connAndError {
		if r.err != nil {
			errs = append(errs, r.err)
		}

		if r.cam != nil {
			log.Info("Connected successfully to camera: [%s]", r.cam.Title())
			s.cameras = append(s.cameras, r.cam)
		}
	}
	return errs
}

func connectToCamera(ctx context.Context, title, addr string, sett camera.Settings) (camera.Connection, error) {
	log.Info("Connecting to camera: [%s]...", title)
	return camera.ConnectWithCancel(ctx, title, addr, sett)
}

func (s *server) LoadConfiguration() error {
	config, err := config.Load()
	if err != nil {
		return err
	}

	s.config = config
	return nil
}

func (s *server) shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, cam := range s.cameras {
		log.Warn("Closing camera connection: [%s]...", cam.Title())
		cam.Close()
	}
	close(s.shutdownDone)
}

func (s *server) Shutdown() chan interface{} {
	s.shutdown()
	return s.shutdownDone
}
