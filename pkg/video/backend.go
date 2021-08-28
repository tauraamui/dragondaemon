package video

import (
	"context"

	"github.com/spf13/afero"
)

var fs = afero.NewOsFs()

type Backend interface {
	Connect(context.Context, string) (Connection, error)
	NewFrame() Frame
	NewWriter() ClipWriter
}

func DefaultBackend() Backend {
	return &openCVBackend{}
}

func MockBackend() Backend {
	return &mockVideoBackend{}
}
