package process

import (
	"context"
	"fmt"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/broadcast"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

const CAM_SWITCHED_OFF_EVT Event = 0x51
const CAM_SWITCHED_ON_EVT Event = 0x52

type streamConnProccess struct {
	ctx      context.Context
	cancel   context.CancelFunc
	listener *broadcast.Listener
	stopping chan interface{}
	cam      camera.Connection
	dest     chan video.Frame
}

func NewStreamConnProcess(
	l *broadcast.Listener, cam camera.Connection, dest chan video.Frame,
) Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &streamConnProccess{
		ctx: ctx, cancel: cancel,
		listener: l,
		cam:      cam, dest: dest, stopping: make(chan interface{}),
	}
}

func (proc *streamConnProccess) Setup() Process { return proc }

func (proc *streamConnProccess) Start() {
	go proc.run()
}

func (proc *streamConnProccess) run() {
	isOn := true
	for {
		time.Sleep(1 * time.Microsecond)
		select {
		case <-proc.ctx.Done():
			close(proc.stopping)
			return
		case msg := <-proc.listener.Ch:
			if e, ok := msg.(Event); ok {
				if e == CAM_SWITCHED_OFF_EVT {
					isOn = false
				} else if e == CAM_SWITCHED_ON_EVT {
					isOn = true
				}
			}
		default:
			if isOn {
				stream(proc.cam, proc.dest)
			}
		}
	}
}

func stream(cam camera.Connection, frames chan video.Frame) {
	if cam.IsOpen() {
		log.Debug("Reading frame from vid stream for camera [%s]", cam.Title())
		frame, err := cam.Read()
		if err != nil {
			log.Error(fmt.Errorf("Unable to retrieve frame: %w. Auto re-connecting is not yet implemented", err).Error())
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
}

func (proc *streamConnProccess) Stop() {
	proc.cancel()
}
func (proc *streamConnProccess) Wait() {
	<-proc.stopping
}
