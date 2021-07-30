package video

import "gocv.io/x/gocv"

type Frame struct {
	mat gocv.Mat
}

func NewFrame() Frame {
	return Frame{mat: gocv.NewMat()}
}

func (f *Frame) Close() {
	f.mat.Close()
}
