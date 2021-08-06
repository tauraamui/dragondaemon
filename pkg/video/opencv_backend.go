package video

import (
	"context"
	"errors"

	"gocv.io/x/gocv"
)

type openCVFrame struct {
	isClosed bool
	mat      gocv.Mat
}

func (frame *openCVFrame) DataRef() interface{} {
	return &frame.mat
}

func (frame *openCVFrame) Close() {
	if !frame.isClosed {
		frame.mat.Close()
		frame.isClosed = true
	}
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

func (b openCVBackend) NewFrame() Frame {
	return &openCVFrame{mat: gocv.NewMat()}
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
	vc, err := openVideoCapture(addr)
	result := openVideoStreamResult{vc: vc, err: err}
	d <- result
}

var openVideoCapture = func(addr string) (*gocv.VideoCapture, error) {
	return gocv.OpenVideoCapture(addr)
}

var readFromVideoConnection = func(vc *gocv.VideoCapture, mat *gocv.Mat) bool {
	return vc.Read(mat)
}

func (c *openCVConnection) Read(frame Frame) error {
	mat, ok := frame.DataRef().(*gocv.Mat)
	if !ok {
		return errors.New("must pass OpenCV frame to OpenCV connection read")
	}
	ok = readFromVideoConnection(c.vc, mat)
	if !ok {
		return errors.New("unable to read from video connection")
	}
	return nil
}

func (c *openCVConnection) Close() error {
	return c.vc.Close()
}
