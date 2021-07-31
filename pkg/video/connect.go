package video

import (
	"context"
	"errors"

	"gocv.io/x/gocv"
)

type Connection interface {
	Read(Frame) error
	Close() error
}

type connection struct {
	// will eventually hide this behind an interface
	vc *gocv.VideoCapture
}

func (c *connection) connect(cancel context.Context, addr string) error {
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

func (c *connection) Read(frame Frame) error {
	ok := c.vc.Read(&frame.mat)
	if !ok {
		return errors.New("unable to read from video connection")
	}
	return nil
}

func (c *connection) Close() error {
	return c.vc.Close()
}

func connect(cancel context.Context, addr string) (Connection, error) {
	conn := connection{}
	err := conn.connect(cancel, addr)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func Connect(addr string) (Connection, error) {
	return connect(context.Background(), addr)
}

func ConnectWithCancel(cancel context.Context, addr string) (Connection, error) {
	return connect(cancel, addr)
}
