package videoframe

type FrameDimension struct {
	W, H int
}

type Frame interface {
	DataRef() interface{}
	Dimensions() FrameDimension
	Close()
}
