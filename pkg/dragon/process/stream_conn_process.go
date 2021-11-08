package process

import (
	"context"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/broadcast"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
	"github.com/tauraamui/xerror"
)

const CAM_SWITCHED_OFF_EVT Event = 0x51
const CAM_SWITCHED_ON_EVT Event = 0x52

type streamConnProccess struct {
	started  chan struct{}
	ctx      context.Context
	cancel   context.CancelFunc
	listener *broadcast.Listener
	stopping chan struct{}
	cam      camera.Connection
	dest     chan videoframe.NoCloser
}

func NewStreamConnProcess(
	l *broadcast.Listener, cam camera.Connection, dest chan videoframe.NoCloser,
) Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &streamConnProccess{
		started: make(chan struct{}),
		ctx:     ctx, cancel: cancel,
		listener: l,
		cam:      cam, dest: dest, stopping: make(chan struct{}),
	}
}

func (proc *streamConnProccess) Setup() Process { return proc }

func (proc *streamConnProccess) Start() <-chan struct{} {
	go run(proc.ctx, proc.cam, proc.dest, *proc.listener, proc.started, proc.stopping)
	return proc.started
}

func run(ctx context.Context, cam camera.Connection, d chan videoframe.NoCloser, l broadcast.Listener, s, stopping chan struct{}) {
	isOn := true
	started := false
	for {
		time.Sleep(1 * time.Microsecond)
		if !started {
			close(s)
			started = true
		}
		select {
		case <-ctx.Done():
			close(stopping)
			return
		case msg := <-l.Ch:
			if e, ok := msg.(Event); ok {
				if e == CAM_SWITCHED_OFF_EVT {
					isOn = false
				}
				if e == CAM_SWITCHED_ON_EVT {
					isOn = true
				}
			}
		default:
			if cam.IsOpen() && isOn {
				stream(cam, d)
			}
		}
	}
}

func stream(cam camera.Connection, frames chan videoframe.NoCloser) {
	log.Debug("Reading frame from vid stream for camera [%s]", cam.Title())
	frame, err := cam.Read()
	if err != nil {
		log.Error(xerror.Errorf("Unable to retrieve frame: %w. Auto re-connecting is not yet implemented", err).Error())
		return
	}
	select {
	case frames <- frame:
		log.Debug("Sending frame from cam to buffer...")
	default:
		frame.Close()
		log.Debug("Buffer full...")
	}
}

func (proc *streamConnProccess) Stop() {
	proc.cancel()
}
func (proc *streamConnProccess) Wait() {
	<-proc.stopping
}
