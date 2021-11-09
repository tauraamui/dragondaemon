package videoclip_test

import (
	"testing"

	"github.com/tauraamui/dragondaemon/pkg/video/videobackend"
	"github.com/tauraamui/dragondaemon/pkg/video/videoclip"
)

func BenchmarkClipAppendFrame(b *testing.B) {
	backend := videobackend.Mock()
	clip := videoclip.New("", 10)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		clip.AppendFrame(backend.NewFrame())
	}
	b.StopTimer()
}
