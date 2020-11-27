package media

import (
	"sync"
	"sync/atomic"

	"gocv.io/x/gocv"
)

type Connection struct {
	inShutdown int32
	vc         *gocv.VideoCapture
	mu         sync.Mutex
	window     *gocv.Window
}

func NewConnection(vc *gocv.VideoCapture) *Connection {
	return &Connection{
		vc: vc,
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

func (c *Connection) Close() error {
	atomic.StoreInt32(&c.inShutdown, 1)
	c.window.Close()
	return c.vc.Close()
}
