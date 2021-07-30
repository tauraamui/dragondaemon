package video

import (
	"errors"

	"gocv.io/x/gocv"
)

type Connection interface {
	Read(Frame) error
	Close() error
}

type connection struct {
	vc *gocv.VideoCapture
}

func (c *connection) connect(addr string) error {
	vc, err := gocv.OpenVideoCapture(addr)
	if err != nil {
		return err
	}
	c.vc = vc
	return nil
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

type connector struct {
	Cancel <-chan interface{}
}

func (c connector) connect(addr string) (Connection, error) {
	conn := connection{}
	err := conn.connect(addr)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func Connect(addr string) (Connection, error) {
	c := connector{}
	return c.connect(addr)
}
