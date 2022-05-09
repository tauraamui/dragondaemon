package videobackend

import (
	"context"

	"github.com/spf13/afero"
	"github.com/tauraamui/dragondaemon/pkg/video/videoclip"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
)

var fs = afero.NewOsFs()

type Connection interface {
	UUID() string
	Read(videoframe.Frame) error
	IsOpen() bool
	Close() error
}

type Backend interface {
	Connect(context.Context, string) (Connection, error)
	NewFrame() videoframe.Frame
	NewFrameFromBytes([]byte) (videoframe.Frame, error)
	NewWriter() videoclip.Writer
}

func Default() Backend {
	return OpenCV()
}

func OpenCV() Backend {
	return &openCVBackend{}
}

func Mock() Backend {
	return &mockVideoBackend{}
}

func Resolve(t string) Backend {
	switch t {
	case "mock":
		return Mock()
	default:
		return Default()
	}
}
