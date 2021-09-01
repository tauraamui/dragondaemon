package process

import (
	"context"
	"time"

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
			clip := makeClip(proc.ctx, proc.frames, proc.framesPerClip, proc.persistLoc)
			if clip != nil {
				proc.dest <- clip
			}
		}
	}
}

func makeClip(ctx context.Context, frames chan video.Frame, count int, persistLoc string) video.Clip {
	clip := video.NewClip(persistLoc, count)
	i := 0
	for f := range frames {
		select {
		case <-ctx.Done():
			clip.Close()
			return nil
		default:
			if i >= count {
				return clip
			}
			clip.AppendFrame(f)
			i++
		}
	}
	return nil
}

func (proc *generateClipProcess) Stop() {
	proc.cancel()
}
func (proc *generateClipProcess) Wait() {
	<-proc.stopping
}
