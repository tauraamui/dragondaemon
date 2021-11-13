package process

import (
	"context"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/broadcast"
	"github.com/tauraamui/dragondaemon/pkg/video/videoclip"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
)

type generateClipProcess struct {
	started       chan struct{}
	ctx           context.Context
	cancel        context.CancelFunc
	listener      *broadcast.Listener
	stopping      chan struct{}
	framesPerClip int
	frames        chan videoframe.NoCloser
	dest          chan videoclip.NoCloser
	persistLoc    string
}

func NewGenerateClipProcess(
	listener *broadcast.Listener, frames chan videoframe.NoCloser, dest chan videoclip.NoCloser, framesPerClip int, persistLoc string,
) Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &generateClipProcess{
		started: make(chan struct{}),
		ctx:     ctx, cancel: cancel,
		listener: listener,
		frames:   frames, dest: dest,
		framesPerClip: framesPerClip,
		persistLoc:    persistLoc,
		stopping:      make(chan struct{}),
	}
}

func (proc *generateClipProcess) Setup() Process { return proc }

func (proc *generateClipProcess) Start() <-chan struct{} {
	go proc.run()
	return proc.started
}

func (proc *generateClipProcess) run() {
	started := false
	for {
		time.Sleep(1 * time.Microsecond)
		if !started {
			close(proc.started)
			started = true
		}
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

func makeClip(ctx context.Context, listener *broadcast.Listener, frames chan videoframe.NoCloser, count int, persistLoc string) videoclip.NoCloser {
	clip := videoclip.New(persistLoc, count)
	i := 0
	for {
		time.Sleep(1 * time.Microsecond)
		select {
		case <-ctx.Done():
			// TODO(tauraamui): this shouldn't do this right? we should just return the clip here
			clip.Close()
			return nil
			// TODO: this will only halt once for the current
			// clip being written to at the moment of the state
			// change, after that this will continue to generate
			// clips, the assumption being that the frames from the
			// stream process will have stopped being sent.
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

func (proc *generateClipProcess) Stop() <-chan struct{} {
	proc.listener.Close()
	proc.cancel()
	return proc.wait()
}

func (proc *generateClipProcess) Wait() {
	<-proc.wait()
}

func (proc *generateClipProcess) wait() <-chan struct{} {
	return proc.stopping
}
