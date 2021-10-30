package video_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

func TestVideoBackendDefaultBackend(t *testing.T) {
	is := is.New(t)
	is.True(video.DefaultBackend() != nil)
}
