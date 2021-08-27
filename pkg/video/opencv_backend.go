package video

import (
	"context"
	"errors"
	"sync"

	"gocv.io/x/gocv"
)

type openCVFrame struct {
	isClosed bool
	mat      gocv.Mat
}

func (frame *openCVFrame) DataRef() interface{} {
	return &frame.mat
}

func (frame *openCVFrame) Dimensions() (int, int) {
	return frame.mat.Cols(), frame.mat.Rows()
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

func (b *openCVBackend) Connect(cancel context.Context, addr string) (Connection, error) {
	err := b.conn.connect(cancel, addr)
	if err != nil {
		return nil, err
	}

	return &b.conn, nil
}

func (b *openCVBackend) NewFrame() Frame {
	return &openCVFrame{mat: gocv.NewMat()}
}

func (b *openCVBackend) NewWriter() ClipWriter {
	return &openCVClipWriter{
		onWriteInitDone: false,
	}
}

const codec = "avc1.4d001e"

type openCVClipWriter struct {
	onWriteInitDone bool
	vw              *gocv.VideoWriter
	clip            Clip
}

func (w *openCVClipWriter) init(clip Clip) error {
	w.clip = clip
	width, height := clip.FrameDimensions()
	vw, err := gocv.VideoWriterFile(
		clip.FileName(), codec, float64(clip.FPS()), width, height, true,
	)
	if err != nil {
		return err
	}
	w.vw = vw
	return nil
}

func (w *openCVClipWriter) reset() {
	w.vw.Close()
	w.vw = nil
}

func (w *openCVClipWriter) Write(clip Clip) error {
	w.init(clip)
	defer w.reset()
	for _, frame := range clip.GetFrames() {
		if err := w.writeFrame(frame); err != nil {
			return err
		}
	}
	return nil
}

func (w *openCVClipWriter) writeFrame(frame Frame) error {
	mat, ok := frame.DataRef().(*gocv.Mat)
	if !ok {
		return errors.New("must pass OpenCV frame to OpenCV writer")
	}
	return w.vw.Write(*mat)
}

type openCVConnection struct {
	mu     sync.Mutex
	isOpen bool
	vc     *gocv.VideoCapture
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
		c.isOpen = true
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
	if vc.IsOpened() {
		return vc.Read(mat)
	}
	return false
}

func (c *openCVConnection) Read(frame Frame) error {
	mat, ok := frame.DataRef().(*gocv.Mat)
	if !ok {
		return errors.New("must pass OpenCV frame to OpenCV connection read")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	ok = readFromVideoConnection(c.vc, mat)
	if !ok {
		return errors.New("unable to read from video connection")
	}
	return nil
}

func (c *openCVConnection) IsOpen() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isOpen {
		return c.vc.IsOpened()
	}
	return false
}

func (c *openCVConnection) Close() error {
	c.mu.Lock()
	c.isOpen = false
	c.mu.Unlock()
	return c.vc.Close()
}
