package camera

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/video"
	"github.com/tauraamui/dragondaemon/pkg/video/videobackend"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
	"github.com/tauraamui/xerror"
)

type Connection interface {
	Reader
	IsOpen
	UUID() string
	Title() string
	PersistLocation() string
	FullPersistLocation() string
	MaxClipAgeDays() int
	FPS() int
	Schedule() schedule.Schedule
	SPC() int
	IsClosing() bool
	Close() error
}

type IsOpen interface {
	IsOpen() bool
}

type Reader interface {
	Read() (videoframe.Frame, error)
}

type connection struct {
	uuid      string
	title     string
	sett      Settings
	backend   videobackend.Backend
	mu        sync.Mutex
	isClosing bool
	vc        videobackend.Connection
}

func (c *connection) UUID() string {
	return c.uuid
}

// TODO(tauraamui): make return typed error and frame
func (c *connection) Read() (videoframe.Frame, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	frame := c.backend.NewFrame()
	if err := c.vc.Read(frame); err != nil {
		return nil, xerror.Errorf("unable to read frame from connection: %w", err)
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

func (c *connection) Schedule() schedule.Schedule {
	if c.sett.Schedule == nil {
		c.sett.Schedule = schedule.NewSchedule(schedule.Week{})
	}
	return c.sett.Schedule
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

func connect(ctx context.Context, title, addr string, settings Settings, backend videobackend.Backend) (Connection, error) {
	vc, err := video.ConnectWithCancel(ctx, addr, backend)
	if err != nil {
		return nil, xerror.Errorf("Unable to connect to camera [%s]: %w", title, err)
	}
	return &connection{
		uuid:    vc.UUID(),
		backend: backend,
		title:   title,
		vc:      vc,
		sett:    settings,
	}, nil
}

func Connect(title, addr string, settings Settings, backend videobackend.Backend) (Connection, error) {
	return connect(context.Background(), title, addr, settings, backend)
}

func ConnectWithCancel(cancel context.Context, title, addr string, settings Settings, backend videobackend.Backend) (Connection, error) {
	return connect(cancel, title, addr, settings, backend)
}
