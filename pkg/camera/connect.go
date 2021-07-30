package camera

import (
	"context"

	"github.com/tauraamui/dragondaemon/pkg/video"
)

type Connection interface {
	Read()
	Close()
}

type connection struct {
	vc video.Connection
	f  video.Frame
}

func (c *connection) Read() {
	c.f = video.NewFrame()
	c.vc.Read(c.f)
}

func (c *connection) Close() {
	c.f.Close()
}

type connector struct {
	Cancel <-chan interface{}
}

func (c connector) connect(ctx context.Context, title, addr string, settings Settings) (Connection, error) {
	vc, err := video.Connect(addr)
	if err != nil {
		return nil, err
	}
	return &connection{
		vc: vc,
	}, nil
}

func Connect(title, addr string, settings Settings) (Connection, error) {
	var connector = connector{}
	return connector.connect(context.Background(), title, addr, settings)
}
