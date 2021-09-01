package process

import (
	"context"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type generateClipProcess struct {
	ctx           context.Context
	cancel        context.CancelFunc
	stopping      chan interface{}
	framesPerClip int
	frames        chan video.Frame
	dest          chan video.Clip
	persistLoc    string
}

func NewGenerateClipProcess(frames chan video.Frame, dest chan video.Clip, framesPerClip int, persistLoc string) Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &generateClipProcess{
		ctx: ctx, cancel: cancel, frames: frames, dest: dest, framesPerClip: framesPerClip, persistLoc: persistLoc, stopping: make(chan interface{}),
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
			proc.dest <- makeClip(proc.frames, proc.framesPerClip, proc.persistLoc)
			log.Debug("pending reading frame from stream")
		}
	}
}

func makeClip(frames chan video.Frame, count int, persistLoc string) video.Clip {
	clip := video.NewClip(persistLoc, count)
	i := 0
	for f := range frames {
		if i >= count {
			break
		}
		clip.AppendFrame(f)
		i++
	}
	return clip
}

func (proc *generateClipProcess) Stop() {}
func (proc *generateClipProcess) Wait() {}
