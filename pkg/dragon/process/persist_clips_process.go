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
			persistClip(<-proc.clips, proc.writer)
		}
	}
}

func persistClip(clip video.Clip, writer video.ClipWriter) {
	writer.Write(clip)
}

func (proc *persistClipProcess) Stop() {}
func (proc *persistClipProcess) Wait() {}
