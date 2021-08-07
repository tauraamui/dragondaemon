package camera

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type Connection interface {
	UUID() string
	Read() video.Frame
	Title() string
	IsOpen() bool
	IsClosing() bool
	Close() error
}

type connection struct {
	uuid      string
	backend   video.Backend
	title     string
	sett      Settings
	mu        sync.Mutex
	isClosing bool
	vc        video.Connection
}

func (c *connection) UUID() string {
	return c.uuid
}

func (c *connection) Read() video.Frame {
	c.mu.Lock()
	defer c.mu.Unlock()
	frame := c.backend.NewFrame()
	c.vc.Read(frame)
	return frame
}

func (c *connection) Title() string {
	return c.title
}

func (c *connection) IsOpen() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.vc.IsOpen()
}

func (c *connection) IsClosing() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isClosing
}

func (c *connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isClosing = true
	return c.vc.Close()
}

func connect(ctx context.Context, title, addr string, settings Settings, backend video.Backend) (Connection, error) {
	vc, err := video.ConnectWithCancel(ctx, addr, backend)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to camera [%s]: %w", title, err)
	}
	return &connection{
		uuid:    uuid.NewString(),
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
