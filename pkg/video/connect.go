package video

import (
	"context"
)

type Connection interface {
	UUID() string
	Read(Frame) error
	IsOpen() bool
	Close() error
}

func connect(cancel context.Context, addr string, backend Backend) (Connection, error) {
	return backend.Connect(cancel, addr)
}

func Connect(addr string, backend Backend) (Connection, error) {
	return connect(context.Background(), addr, backend)
}

func ConnectWithCancel(cancel context.Context, addr string, backend Backend) (Connection, error) {
	return connect(cancel, addr, backend)
}
