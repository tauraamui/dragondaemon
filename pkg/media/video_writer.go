package media

import (
	"gocv.io/x/gocv"
)

type VideoWriteable interface {
	SetP(*gocv.VideoWriter)
	IsOpened() bool
	Write(gocv.Mat) error
	Close() error
}

type videoWriter struct {
	p *gocv.VideoWriter
}

var openVideoWriter = func(
	fileName string,
	codec string,
	fps float64,
	frameWidth int,
	frameHeight int,
	mock bool,
) (VideoWriteable, error) {
	if mock {
		return &mockVideoWriter{}, nil
	}

	vw, err := gocv.VideoWriterFile(fileName, codec, fps, frameWidth, frameHeight, true)
	if err != nil {
		return nil, err
	}

	return &videoWriter{
		p: vw,
	}, err
}

// SetP updates the internal pointer to the video capture instance.
func (vw *videoWriter) SetP(w *gocv.VideoWriter) {
	vw.p = w
}

func (vw *videoWriter) IsOpened() bool {
	return vw.p.IsOpened()
}

func (vw *videoWriter) Write(m gocv.Mat) error {
	return vw.p.Write(m)
}

func (vw *videoWriter) Close() error {
	return vw.p.Close()
}

type mockVideoWriter struct {
	initialised bool
}

// SetP doesn't do anything it exists to satisfy videoWriteable interface
func (mvw *mockVideoWriter) SetP(_ *gocv.VideoWriter) {}

// IsOpened always returns true.
func (mvw *mockVideoWriter) IsOpened() bool {
	return true
}

func (mvw *mockVideoWriter) Write(m gocv.Mat) error {
	return nil
}

func (mvw *mockVideoWriter) Close() error {
	mvw.initialised = false
	return nil
}
