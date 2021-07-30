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

func (c *connection) connect(device interface{}) error {
	vc, err := gocv.OpenVideoCapture(device)
	if err != nil {
		return err
	}
	c.vc = vc
	return nil
}

func (c *connection) Read(frame Frame) error {
	ok := c.vc.Read(frame.mat)
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

func (c connector) connect(device interface{}) (Connection, error) {
	conn := connection{}
	err := conn.connect(device)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func Connect(device interface{}) (Connection, error) {
	c := connector{}
	return c.connect(device)
}
