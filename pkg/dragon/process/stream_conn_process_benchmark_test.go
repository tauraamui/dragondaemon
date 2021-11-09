package process_test

import (
	"context"
	"testing"

	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/video/videobackend"
)

func BenchmarkRunningStreamConnProcess(b *testing.B) {
	backend := videobackend.Mock()
	conn, err := backend.Connect(context.Background(), "fake-addr")
	if err != nil {
		b.Fatal(err)
	}
	camera.Connect("fake-cam", "fake-addr", camera.Settings{}, backend)
}
