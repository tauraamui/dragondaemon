package video

import (
	"context"
	"errors"

	"gocv.io/x/gocv"
)

type openCVFrame struct {
	mat gocv.Mat
}

func (frame openCVFrame) DataRef() interface{} {
	return &frame.mat
}

func (frame openCVFrame) Close() {
	frame.mat.Close()
}

type OpenCVBackend struct {
	conn openCVConnection
}

func (b OpenCVBackend) Connect(cancel context.Context, addr string) (Connection, error) {
	err := b.conn.connect(cancel, addr)
	if err != nil {
		return &openCVConnection{}, err
	}

	return &b.conn, nil
}

func (b OpenCVBackend) NewFrame() Frame {
	return openCVFrame{mat: gocv.NewMat()}
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
	mat, ok := frame.DataRef().(*gocv.Mat)
	if !ok {
		return errors.New("must pass OpenCV frame to OpenCV connection read")
	}
	ok = c.vc.Read(mat)
	if !ok {
		return errors.New("unable to read from video connection")
	}
	return nil
}

func (c *openCVConnection) Close() error {
	return c.vc.Close()
}
