package process

import (
	"context"
	"fmt"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

const PROC_CAM_SWITCHED_OFF = 0x51

type streamConnProccess struct {
	ctx       context.Context
	cancel    context.CancelFunc
	callbacks map[event]func()
	stopping  chan interface{}
	cam       camera.Connection
	isOff     bool
	wasOff    bool
	dest      chan video.Frame
}

func NewStreamConnProcess(cam camera.Connection, dest chan video.Frame) Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &streamConnProccess{
		ctx: ctx, cancel: cancel, callbacks: map[event]func(){}, cam: cam, dest: dest, stopping: make(chan interface{}),
	}
}

func (proc *streamConnProccess) Setup() {}

func (proc *streamConnProccess) RegisterCallback(code event, callback func()) error {
	proc.callbacks[code] = callback
	return nil
}

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
			println("schedule", proc.cam.Schedule())
			proc.isOff = !proc.cam.Schedule().IsOn(schedule.Now())
			if proc.isOff && !proc.wasOff {
				if switchedOffCallback := proc.callbacks[PROC_CAM_SWITCHED_OFF]; switchedOffCallback != nil {
					switchedOffCallback()
				}
				proc.wasOff = true
				continue
			}

			if !proc.isOff {
				proc.wasOff = false
			}

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
