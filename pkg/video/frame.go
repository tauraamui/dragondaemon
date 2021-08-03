package video

// type Frame struct {
// 	mat gocv.Mat
// }

type Frame interface {
	DataRef() interface{}
	Close()
}
