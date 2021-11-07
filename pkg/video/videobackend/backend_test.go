package videobackend_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/tauraamui/dragondaemon/pkg/video/videobackend"
)

func TestVideoBackendDefaultBackend(t *testing.T) {
	is := is.New(t)
	is.True(videobackend.Default() != nil)
}
