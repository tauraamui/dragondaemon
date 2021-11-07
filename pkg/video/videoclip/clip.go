package videoclip

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
	"github.com/tauraamui/xerror"
)

type Clip interface {
	NoCloser
	Closer
}

type NoCloser interface {
	AppendFrame(videoframe.Frame)
	GetFrames() []videoframe.Frame
	FrameDimensions() (videoframe.Dimensions, error)
	FPS() int
	RootPath() string
	FileName() string
}

type Closer interface {
	Close()
}

const DATE_FORMAT = "2006-01-02"
const DATE_AND_TIME_FORMAT = "2006-01-02 15.04.05"

var Timestamp = func() time.Time {
	return time.Now()
}

func New(ploc string, fps int) Clip {
	return &clip{
		timestamp:           Timestamp(),
		fps:                 fps,
		rootPersistLocation: ploc,
		isClosed:            false,
	}
}

type clip struct {
	timestamp           time.Time
	rootPersistLocation string
	fps                 int
	mu                  sync.Mutex
	isClosed            bool
	frames              []videoframe.Frame
}

func (c *clip) AppendFrame(f videoframe.Frame) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		log.Fatal("cannot append frame to closed clip")
	}
	c.frames = append(c.frames, f)
}

func (c *clip) FrameDimensions() (videoframe.Dimensions, error) {
	if len(c.frames) == 0 {
		return videoframe.Dimensions{}, xerror.New("unable to resolve clip's footage dimensions")
	}
	return c.frames[0].Dimensions(), nil
}

func (c *clip) FPS() int {
	return c.fps
}

func (c *clip) RootPath() string {
	return filepath.Join(c.rootPersistLocation, c.timestamp.Format(DATE_FORMAT))
}

func (c *clip) FileName() string {
	return filepath.FromSlash(
		fmt.Sprintf(
			"%s/%s/%s.mp4",
			c.rootPersistLocation,
			c.timestamp.Format(DATE_FORMAT),
			c.timestamp.Format(DATE_AND_TIME_FORMAT)),
	)
}

func (c *clip) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, frame := range c.frames {
		frame.Close()
	}

	c.isClosed = true
}

func (c *clip) GetFrames() []videoframe.Frame {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.frames
}
