package video

import (
	"sync"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/log"
)

type Clip interface {
	AppendFrame(Frame)
	FrameDimensions() (int, int)
	FPS() int
	PersistLocation() string
	Close()
}

var Timestamp = func() time.Time {
	return time.Now()
}

func NewClip(ploc string) *clip {
	return &clip{
		timestamp:       Timestamp(),
		persistLocation: ploc,
		isClosed:        false,
	}
}

type clip struct {
	timestamp       time.Time
	persistLocation string
	fps             int
	mu              sync.Mutex
	isClosed        bool
	frames          []Frame
}

func (c *clip) AppendFrame(f Frame) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		log.Fatal("cannot append frame to closed clip")
	}
	c.frames = append(c.frames, f)
}

func (c *clip) FrameDimension() (int, int) {
	if len(c.frames) == 0 {
		return 0, 0
	}
	return c.frames[0].Dimensions()
}

func (c *clip) FPS() int {
	return c.fps
}

func (c *clip) PersistLocation() string {
	return c.persistLocation
}

func (c *clip) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, frame := range c.frames {
		frame.Close()
	}

	c.isClosed = true
}
