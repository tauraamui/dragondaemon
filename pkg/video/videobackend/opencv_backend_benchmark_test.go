package videobackend_test

import (
	"context"
	"os"
	"testing"

	"github.com/tauraamui/dragondaemon/internal/videotest"
	"github.com/tauraamui/dragondaemon/pkg/video/videobackend"
)

func BenchmarkOpenCVBackendConnect(b *testing.B) {
	mp4FilePath, err := videotest.RestoreMp4File()
	if err != nil {
		b.Fatal(err)
	}
	defer func() { os.Remove(mp4FilePath) }()

	backend := videobackend.OpenCV()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		conn, err := backend.Connect(context.Background(), mp4FilePath)
		if conn == nil || err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}
