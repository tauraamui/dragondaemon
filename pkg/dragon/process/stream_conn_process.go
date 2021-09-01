package process

import (
	"context"
	"fmt"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type streamConnProccess struct {
	ctx      context.Context
	cancel   context.CancelFunc
	stopping chan interface{}
	cam      camera.Connection
	dest     chan video.Frame
}

func NewStreamConnProcess(cam camera.Connection, dest chan video.Frame) Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &streamConnProccess{
		ctx: ctx, cancel: cancel, cam: cam, dest: dest, stopping: make(chan interface{}),
	}
}

func (proc *streamConnProccess) Setup() {}
func (proc *streamConnProccess) Start() {
	go proc.run()
}

func (proc *streamConnProccess) run() {
	for {
		time.Sleep(1 * time.Microsecond)
		select {
		case <-proc.ctx.Done():
			close(proc.stopping)
			return
		default:
			stream(proc.cam, proc.dest)
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