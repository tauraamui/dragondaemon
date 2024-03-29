package process

import (
	"testing"

	"github.com/tauraamui/dragondaemon/pkg/broadcast"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/mocks"
	"github.com/tauraamui/dragondaemon/pkg/video/videobackend"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
)

func BenchmarkReadingFramesFromInfinateLowSizeFrameProducer(b *testing.B) {
	b.ResetTimer()
	b.StopTimer()
	dest := make(chan videoframe.NoCloser)
	stopReads := make(chan struct{})
	go func(frames <-chan videoframe.NoCloser, stop <-chan struct{}) {
		for {
			select {
			case <-stop:
				return
			case <-frames:
			default:
				continue
			}
		}
	}(dest, stopReads)
	conn := mocks.NewCamConn(mocks.Options{UntrackedFrames: true})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		stream(conn.Title(), conn, dest)
	}
	b.StopTimer()
	close(stopReads)
	close(dest)
}

func BenchmarkStreamConnProcessReading10000LowSizeFrames(b *testing.B) {
	b.StopTimer()
	b.ResetTimer()

	const maxFrames uint = 10000

	readFrames := make(chan videoframe.NoCloser, 3)
	conn := mocks.NewCamConn(mocks.Options{UntrackedFrames: true, IsOpen: true})
	proc := NewStreamConnProcess(broadcast.New(0).Listen(), "testCam", conn, readFrames)

	proc.Setup().Start()

	b.StartTimer()
	var count uint
procLoop:
	for {
		select {
		case <-readFrames:
			count++
		default:
			if count >= maxFrames {
				break procLoop
			}
		}
	}
	b.StopTimer()

	<-proc.Stop()
}

func BenchmarkStreamConnProcessReading30FramesFromMockVideoStream(b *testing.B) {
	b.StopTimer()
	b.ResetTimer()
	const maxFrames uint = 30

	readFrames := make(chan videoframe.NoCloser, 3)
	conn, err := camera.Connect("mock", "", camera.Settings{}, videobackend.Mock())
	if err != nil {
		b.Fatal("unable to open mock connection: %w", err)
	}

	proc := NewStreamConnProcess(broadcast.New(0).Listen(), "testCam", conn, readFrames)

	proc.Setup().Start()

	b.StartTimer()
	var count uint
procLoop:
	for {
		select {
		case <-readFrames:
			count++
		default:
			if count >= maxFrames {
				break procLoop
			}
		}
	}
	b.StopTimer()

	<-proc.Stop()
}
