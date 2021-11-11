package process

import (
	"testing"

	"github.com/tauraamui/dragondaemon/pkg/mocks"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
)

func BenchmarkReadingFramesFromMockBackendConnectionStream(b *testing.B) {
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
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		stream(conn.Title(), conn, dest)
	}
	b.StopTimer()
	close(stopReads)
	close(dest)
}
