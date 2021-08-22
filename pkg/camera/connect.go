package camera

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

type Connection interface {
	UUID() string
	Title() string
	PersistLocation() string
	FullPersistLocation() string
	MaxClipAgeDays() int
	FPS() int
	SPC() int
	Read() (video.Frame, error)
	IsOpen() bool
	IsClosing() bool
	Close() error
}

type connection struct {
	uuid      string
	title     string
	sett      Settings
	backend   video.Backend
	mu        sync.Mutex
	isClosing bool
	vc        video.Connection
}

func (c *connection) UUID() string {
	return c.uuid
}

// TODO(tauraamui): make return typed error and frame
func (c *connection) Read() (video.Frame, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	frame := c.backend.NewFrame()
	if err := c.vc.Read(frame); err != nil {
		return nil, fmt.Errorf("unable to read frame from connection: %w", err)
	}
	return frame, nil
}

func (c *connection) Title() string {
	return c.title
}

func (c *connection) PersistLocation() string {
	return c.sett.PersistLocation
}

func (c *connection) FullPersistLocation() string {
	return filepath.FromSlash(fmt.Sprintf("%s/%s", c.PersistLocation(), c.Title()))
}

func (c *connection) MaxClipAgeDays() int {
	return c.sett.MaxClipAgeDays
}

func (c *connection) FPS() int {
	return c.sett.FPS
}

func (c *connection) SPC() int {
	return c.sett.SecondsPerClip
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
