package process

import (
	"context"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/video"
)

type persistClipProcess struct {
	ctx      context.Context
	cancel   context.CancelFunc
	stopping chan interface{}
	clips    chan video.Clip
	writer   video.ClipWriter
}

func NewPersistClipProcess(clips chan video.Clip, writer video.ClipWriter) Process {
	ctx, cancel := context.WithCancel(context.Background())
	return &persistClipProcess{
		ctx: ctx, cancel: cancel, clips: clips, writer: writer, stopping: make(chan interface{}),
	}
}

func (proc *persistClipProcess) Setup() {}

func (proc *persistClipProcess) Start() {
	go proc.run()
}

func (proc *persistClipProcess) run() {
	for {
		time.Sleep(1 * time.Microsecond)
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
