package video

import (
	"context"

	"github.com/tauraamui/dragondaemon/pkg/video/videobackend"
)

func connect(cancel context.Context, addr string, backend videobackend.Backend) (videobackend.Connection, error) {
	return backend.Connect(cancel, addr)
}

func Connect(addr string, backend videobackend.Backend) (videobackend.Connection, error) {
	return connect(context.Background(), addr, backend)
}

func ConnectWithCancel(cancel context.Context, addr string, backend videobackend.Backend) (videobackend.Connection, error) {
	return connect(cancel, addr, backend)
}
