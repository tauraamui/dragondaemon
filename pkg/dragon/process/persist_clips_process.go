package process

import (
	"context"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/video"
)

type persistClipProcess struct {
	started  chan interface{}
	ctx      context.Context
	cancel   context.CancelFunc
	stopping chan interface{}
	clips    chan video.Clip
	writer   video.ClipWriter
}

func NewPersistClipProcess(clips chan video.Clip, writer video.ClipWriter) Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &persistClipProcess{
		started: make(chan interface{}), ctx: ctx, cancel: cancel, clips: clips, writer: writer, stopping: make(chan interface{}),
	}
}

func (proc *persistClipProcess) Setup() Process { return proc }

func (proc *persistClipProcess) Start() <-chan interface{} {
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
				proc.writer.Write(clip)
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
