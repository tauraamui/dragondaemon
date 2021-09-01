package process

import (
	"context"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type generateClipProcess struct {
	ctx      context.Context
	cancel   context.CancelFunc
	stopping chan interface{}
	frames   chan video.Frame
	dest     chan video.Clip
}

func NewGenerateClipProcess(frames chan video.Frame, dest chan video.Clip, count int) Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &generateClipProcess{
		ctx: ctx, cancel: cancel, frames: frames, dest: dest, stopping: make(chan interface{}),
	}
}

func (proc *generateClipProcess) Setup() {}
func (proc *generateClipProcess) Start() {
	go proc.run()
}

func (proc *generateClipProcess) run() {
	for {
		time.Sleep(1 * time.Microsecond)
		select {
		case <-proc.ctx.Done():
			close(proc.stopping)
			return
		default:
			log.Debug("pending reading frame from stream")
		}
	}
}

func (proc *generateClipProcess) Stop() {}
func (proc *generateClipProcess) Wait() {}
