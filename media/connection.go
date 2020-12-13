package media

import (
	"sync"
	"sync/atomic"

	"gocv.io/x/gocv"
)

type Connection struct {
	inShutdown      int32
	title           string
	persistLocation string
	secondsPerClip  int
	vc              *gocv.VideoCapture
	mu              sync.Mutex
	window          *gocv.Window
}

func NewConnection(
	title string,
	persistLocation string,
	secondsPerClip int,
	vc *gocv.VideoCapture,
) *Connection {
	return &Connection{
		title:           title,
		persistLocation: persistLocation,
		secondsPerClip:  secondsPerClip,
		vc:              vc,
	}
}

func (c *Connection) ShowInWindow(winTitle string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.window != nil {
		c.window.Close()
	}

	c.window = gocv.NewWindow(winTitle)
	img := gocv.NewMat()
	defer img.Close()

	for atomic.LoadInt32(&c.inShutdown) == 0 {
		c.vc.Read(&img)
		if img.Empty() {
			continue
		}
		c.window.IMShow(img)
		c.window.WaitKey(1)
	}
}

func (c *Connection) Title() string {
	return c.title
}

func (c *Connection) Close() error {
	atomic.StoreInt32(&c.inShutdown, 1)
	if c.window != nil {
		c.window.Close()
	}
	return c.vc.Close()
}
