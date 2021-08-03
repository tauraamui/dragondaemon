package video

import (
	"context"
	"errors"

	"gocv.io/x/gocv"
)

type Backend interface {
	Connect(context.Context, string) (Connection, error)
}

func DefaultBackend() Backend {
	return openCVBackend{}
}

type openCVBackend struct {
	conn openCVConnection
}

func (b openCVBackend) Connect(cancel context.Context, addr string) (Connection, error) {
	err := b.conn.connect(cancel, addr)
	if err != nil {
		return &openCVConnection{}, err
	}

	return &b.conn, nil
}

type Connection interface {
	Read(Frame) error
	Close() error
}

type openCVConnection struct {
	// will eventually hide this behind an interface
	vc *gocv.VideoCapture
}

func (c *openCVConnection) connect(cancel context.Context, addr string) error {
	connAndError := make(chan openVideoStreamResult)
	go openVideoStream(addr, connAndError)
	select {
	case r := <-connAndError:
		if r.err != nil {
			return r.err
		}
		c.vc = r.vc
		return nil
	case <-cancel.Done():
		return errors.New("connection cancelled")
	}
}

type openVideoStreamResult struct {
	vc  *gocv.VideoCapture
	err error
}

func openVideoStream(addr string, d chan openVideoStreamResult) {
	vc, err := gocv.OpenVideoCapture(addr)
	result := openVideoStreamResult{vc: vc, err: err}
	d <- result
}

func (c *openCVConnection) Read(frame Frame) error {
	ok := c.vc.Read(&frame.mat)
	if !ok {
		return errors.New("unable to read from video connection")
	}
	return nil
}

func (c *openCVConnection) Close() error {
	return c.vc.Close()
}

func connect(cancel context.Context, addr string, backend Backend) (Connection, error) {
	return backend.Connect(cancel, addr)
}

func Connect(addr string, backend Backend) (Connection, error) {
	return connect(context.Background(), addr, backend)
}

func ConnectWithCancel(cancel context.Context, addr string, backend Backend) (Connection, error) {
	return connect(cancel, addr, backend)
}
