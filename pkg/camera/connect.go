package camera

import (
	"context"
	"fmt"

	"github.com/tauraamui/dragondaemon/pkg/video"
)

type Connection interface {
	Read()
	Title() string
	Close()
}

type connection struct {
	backend video.Backend
	title   string
	sett    Settings
	vc      video.Connection
	f       video.Frame
}

func (c *connection) Read() {
	c.f = c.backend.NewFrame()
	c.vc.Read(c.f)
}

func (c *connection) Title() string {
	return c.title
}

func (c *connection) Close() {
	c.f.Close()
}

func connect(ctx context.Context, title, addr string, settings Settings, backend video.Backend) (Connection, error) {
	vc, err := video.ConnectWithCancel(ctx, addr, backend)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to camera [%s]: %w", title, err)
	}
	return &connection{
		backend: backend,
		title:   title,
		vc:      vc,
		sett:    settings,
	}, nil
}

func Connect(title, addr string, settings Settings, backend video.Backend) (Connection, error) {
	return connect(context.Background(), title, addr, settings, backend)
}

func ConnectWithCancel(cancel context.Context, title, addr string, settings Settings, backend video.Backend) (Connection, error) {
	return connect(cancel, title, addr, settings, backend)
}
