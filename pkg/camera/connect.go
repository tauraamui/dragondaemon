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
}

func (c connection) Read() {
	f := video.NewFrame()
	defer f.Close()
	c.vc.Read(f)
}

func (c connection) Close() {}

type connector struct {
	Cancel <-chan interface{}
}

func (c connector) connect(ctx context.Context, title, addr interface{}, settings Settings) (Connection, error) {
	vc, err := video.Connect(addr)
	if err != nil {
		return nil, err
	}
	return connection{
		vc: vc,
	}, nil
}

func Connect(title, addr interface{}, settings Settings) (Connection, error) {
	var connector = connector{}
	return connector.connect(context.Background(), title, addr, settings)
}
