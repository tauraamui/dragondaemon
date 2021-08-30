package dragon

import (
	"context"
	"sync"

	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/config"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type Server interface {
	Connect() []error
	ConnectWithCancel(context.Context) []error
	LoadConfiguration() error
	SetupProcesses()
	RunProcesses()
	Shutdown() chan interface{}
}

func NewServer(cr config.Resolver, vb video.Backend) Server {
	return &server{
		configResolver: cr,
		videoBackend:   vb,
		shutdownDone:   make(chan interface{}),
	}
}

type server struct {
	configResolver config.Resolver
	videoBackend   video.Backend
	shutdownDone   chan interface{}
	config         configdef.Values
	mu             sync.Mutex
	coreProcesses  []process.Process
	cameras        []camera.Connection
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
	connAndError := make(chan connectResult)
	wg := sync.WaitGroup{}
	wg.Add(len(s.config.Cameras))
	for _, cam := range s.config.Cameras {
		go func(cancel context.Context, wg *sync.WaitGroup, cam configdef.Camera, connAndError chan connectResult) {
			defer wg.Done()
			select {
			case <-cancel.Done():
				return
			default:
				r := connect(cancel, cam, s.videoBackend)
				if r != nil {
					connAndError <- *r
				}
			}
		}(cancel, &wg, cam, connAndError)
	}

	go func(c chan connectResult, wg *sync.WaitGroup) {
		wg.Wait()
		close(c)
	}(connAndError, &wg)

	return s.recieveConnsToTrack(connAndError)
}

func (s *server) recieveConnsToTrack(connAndError chan connectResult) []error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errs []error
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

func connect(cancel context.Context, cam configdef.Camera, backend video.Backend) *connectResult {
	if cam.Disabled {
		log.Warn("Camera [%s] is disabled... skipping...", cam.Title)
		return nil
	}
	settings := camera.Settings{
		DateTimeFormat:  cam.DateTimeFormat,
		DateTimeLabel:   cam.DateTimeLabel,
		FPS:             cam.FPS,
		SecondsPerClip:  cam.SecondsPerClip,
		PersistLocation: cam.PersistLoc,
		MaxClipAgeDays:  cam.MaxClipAgeDays,
		Reolink:         cam.ReolinkAdvanced,
	}

	conn, err := connectToCamera(cancel, cam.Title, cam.Address, settings, backend)
	return &connectResult{
		cam: conn,
		err: err,
	}
}

func connectToCamera(ctx context.Context, title, addr string, sett camera.Settings, backend video.Backend) (camera.Connection, error) {
	log.Info("Connecting to camera: [%s]...", title)
	return camera.ConnectWithCancel(ctx, title, addr, sett, backend)
}

func (s *server) LoadConfiguration() error {
	config, err := s.configResolver.Resolve()
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
	s.shutdownProcesses()
	s.shutdown()
	return s.shutdownDone
}
