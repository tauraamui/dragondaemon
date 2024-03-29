package dragon

import (
	"context"
	"sync"

	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/config"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/configdef"
	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video/videobackend"
	"github.com/tauraamui/xerror"
)

// type Server interface {
// 	Connect() []error
// 	ConnectWithCancel(context.Context) []error
// 	SetupProcesses()
// 	RunProcesses()
// 	Shutdown() <-chan struct{}
// }

func NewServer(cr config.Resolver, vb videobackend.Backend) (*Server, error) {
	c, err := cr.Resolve()
	if err != nil {
		return nil, xerror.Errorf("unable to resolve config: %w", err)
	}

	return &Server{
		config:        c,
		videoBackend:  vb,
		coreProcesses: map[string]process.Process{},
		shutdownDone:  make(chan struct{}),
	}, nil
}

type Server struct {
	runtimeStatsEnabled    bool
	renderRuntimeStatsProc process.Process
	videoBackend           videobackend.Backend
	shutdownDone           chan struct{}
	config                 configdef.Values
	mu                     sync.Mutex
	coreProcesses          map[string]process.Process
	cameras                []camera.Connection
}

func (s *Server) Connect() []error {
	return s.connect(context.Background())
}

func (s *Server) ConnectWithCancel(cancel context.Context) []error {
	return s.connect(cancel)
}

type connectResult struct {
	cam camera.Connection
	err error
}

func (s *Server) connect(cancel context.Context) []error {
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

func (s *Server) recieveConnsToTrack(connAndError chan connectResult) []error {
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

func connect(cancel context.Context, cam configdef.Camera, backend videobackend.Backend) *connectResult {
	if cam.Disabled {
		log.Warn("Camera [%s] is disabled... skipping...", cam.Title)
		return nil
	}
	settings := camera.Settings{
		DateTimeFormat:  cam.DateTimeFormat,
		DateTimeLabel:   cam.DateTimeLabel,
		FPS:             cam.FPS,
		Schedule:        schedule.NewSchedule(cam.Week),
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

func connectToCamera(ctx context.Context, title, addr string, sett camera.Settings, backend videobackend.Backend) (camera.Connection, error) {
	log.Info("Connecting to camera: [%s@%s]...", title, addr)
	return camera.ConnectWithCancel(ctx, title, addr, sett, backend)
}

func (s *Server) shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, cam := range s.cameras {
		log.Warn("Closing camera connection: [%s]...", cam.Title())
		cam.Close()
	}
	close(s.shutdownDone)
}

func (s *Server) Shutdown() <-chan struct{} {
	s.shutdownProcesses()
	s.shutdown()
	return s.shutdownDone
}
