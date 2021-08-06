package video

import (
	"context"
)

type Backend interface {
	Connect(context.Context, string) (Connection, error)
	NewFrame() Frame
}

func DefaultBackend() Backend {
	return &openCVBackend{}
}
