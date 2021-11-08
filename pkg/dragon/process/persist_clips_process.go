package process

import (
	"context"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video/videoclip"
)

type persistClipProcess struct {
	started  chan struct{}
	ctx      context.Context
	cancel   context.CancelFunc
	stopping chan struct{}
	clips    chan videoclip.NoCloser
	writer   videoclip.Writer
}

func NewPersistClipProcess(clips chan videoclip.NoCloser, writer videoclip.Writer) Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &persistClipProcess{
		started: make(chan struct{}), ctx: ctx, cancel: cancel, clips: clips, writer: writer, stopping: make(chan struct{}),
	}
}

func (proc *persistClipProcess) Setup() Process { return proc }

func (proc *persistClipProcess) Start() <-chan struct{} {
	go proc.run()
	return proc.started
}

func (proc *persistClipProcess) run() {
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
			select {
			case clip := <-proc.clips:
				if err := proc.writer.Write(clip); err != nil {
					log.Error(err.Error())
				}
				// clip.Close()
			default:
				continue
			}
		}
	}
}

func (proc *persistClipProcess) Stop() {
	proc.cancel()
}
func (proc *persistClipProcess) Wait() {
	<-proc.stopping
}
