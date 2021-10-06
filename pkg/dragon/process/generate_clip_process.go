package process

import (
	"context"
	"errors"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/video"
)

const PROC_FORCE_DUMP_CURRENT_CLIP = 0x52

type generateClipProcess struct {
	ctx           context.Context
	cancel        context.CancelFunc
	events        chan Event
	stopping      chan interface{}
	framesPerClip int
	frames        chan video.Frame
	dest          chan video.Clip
	persistLoc    string
}

func NewGenerateClipProcess(events chan Event, frames chan video.Frame, dest chan video.Clip, framesPerClip int, persistLoc string) Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &generateClipProcess{
		ctx: ctx, cancel: cancel, events: events, frames: frames, dest: dest, framesPerClip: framesPerClip, persistLoc: persistLoc, stopping: make(chan interface{}),
	}
}

func (proc *generateClipProcess) Setup() {}

func (proc *generateClipProcess) RegisterCallback(code Event, callback func()) error {
	return errors.New("generate clip proc does not support event callbacks")
}

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
			clip := makeClip(proc.ctx, proc.events, proc.frames, proc.framesPerClip, proc.persistLoc)
			if clip != nil {
				proc.dest <- clip
			}
		}
	}
}

func makeClip(ctx context.Context, events chan Event, frames chan video.Frame, count int, persistLoc string) video.Clip {
	clip := video.NewClip(persistLoc, count)
	i := 0
	for f := range frames {
		select {
		case <-ctx.Done():
			// TODO(tauraamui): this shouldn't do this right? we should just return the clip here
			clip.Close()
			return nil
		case e := <-events:
			if e == PROC_FORCE_DUMP_CURRENT_CLIP {
				return clip
			}
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
