package video

import (
	"sync"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/log"
)

type Clip interface {
	AppendFrame(Frame)
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

func (c *clip) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, frame := range c.frames {
		frame.Close()
	}

	c.isClosed = true
}
