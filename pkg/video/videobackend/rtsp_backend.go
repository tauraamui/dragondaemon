package videobackend

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"
)

type rtspBackend struct{}

func (b *rtspBackend) Connect(cancel context.Context, addr string) (Connection, error) {
	return &rtspConnection{}, nil
}

type vidConn interface {
	IsOpened() bool
	Close() error
}

type rtspConnection struct {
	uuid   string
	mu     sync.Mutex
	isOpen bool
}

func (c *rtspConnection) connect(cancel context.Context, addr string) error {
	return errors.New("rtsp backend connection not yet implemented")
}

func (c *rtspConnection) UUID() string {
	if len(c.uuid) == 0 {
		c.uuid = uuid.NewString()
	}
	return c.uuid
}

func (c *rtspConnection) Read(frame videoframe.Frame) error {
	// mat, ok := frame.DataRef().(*gocv.Mat)
	// if !ok {
	// 	return xerror.New("must pass OpenCV frame to OpenCV connection read")
	// }
	// c.mu.Lock()
	// defer c.mu.Unlock()
	// ok = readFromVideoConnection(c.vc, mat)
	// if !ok {
	// 	return xerror.New("unable to read from video connection")
	// }
	return errors.New("reading from rtsp vid conn not yet implemented")
}

func (c *rtspConnection) IsOpen() bool { return false }

func (c *rtspConnection) Close() error { return nil }
