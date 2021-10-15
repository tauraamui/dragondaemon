package process

import (
	"context"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/broadcast"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type generateClipProcess struct {
	ctx           context.Context
	cancel        context.CancelFunc
	listener      *broadcast.Listener
	stopping      chan interface{}
	framesPerClip int
	frames        chan video.Frame
	dest          chan video.Clip
	persistLoc    string
}

func NewGenerateClipProcess(
	listener *broadcast.Listener, frames chan video.Frame, dest chan video.Clip, framesPerClip int, persistLoc string,
) Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &generateClipProcess{
		ctx: ctx, cancel: cancel,
		listener: listener,
		frames:   frames, dest: dest,
		framesPerClip: framesPerClip,
		persistLoc:    persistLoc,
		stopping:      make(chan interface{}),
	}
}

func (proc *generateClipProcess) Setup() Process { return proc }

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
			clip := makeClip(proc.ctx, proc.listener, proc.frames, proc.framesPerClip, proc.persistLoc)
			if clip != nil {
				proc.dest <- clip
			}
		}
	}
}

func makeClip(ctx context.Context, listener *broadcast.Listener, frames chan video.Frame, count int, persistLoc string) video.Clip {
	clip := video.NewClip(persistLoc, count)
	i := 0
	for {
		time.Sleep(1 * time.Microsecond)
		select {
		case <-ctx.Done():
			// TODO(tauraamui): this shouldn't do this right? we should just return the clip here
			clip.Close()
			return nil
		case msg := <-listener.Ch:
			if e, ok := msg.(Event); ok && e == CAM_SWITCHED_OFF_EVT {
				return clip
			}
		case f := <-frames:
			if i >= count {
				return clip
			}
			clip.AppendFrame(f)
			i++
		}
	}
}

func (proc *generateClipProcess) Stop() {
	proc.listener.Close()
	proc.cancel()
}

func (proc *generateClipProcess) Wait() {
	<-proc.stopping
}
