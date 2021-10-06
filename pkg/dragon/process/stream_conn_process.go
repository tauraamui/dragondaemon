package process

import (
	"context"
	"fmt"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/broadcast"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

const CAM_SWITCHED_OFF_EVT Event = 0x51

type streamConnProccess struct {
	ctx         context.Context
	cancel      context.CancelFunc
	broadcaster *broadcast.Broadcaster
	stopping    chan interface{}
	cam         camera.Connection
	isOff       bool
	wasOff      bool
	dest        chan video.Frame
}

func NewStreamConnProcess(broadcaster *broadcast.Broadcaster, cam camera.Connection, dest chan video.Frame) Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &streamConnProccess{
		ctx: ctx, cancel: cancel,
		broadcaster: broadcaster,
		cam:         cam, dest: dest, stopping: make(chan interface{}),
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
			proc.isOff = !proc.cam.Schedule().IsOn(schedule.Time(time.Now()))
			if proc.isOff {
				if !proc.wasOff {
					proc.wasOff = true
					proc.broadcaster.Send(CAM_SWITCHED_OFF_EVT)
				}
				continue
			}

			proc.wasOff = false
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
