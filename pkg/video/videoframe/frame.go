package videoframe

type Dimensions struct {
	W, H int
}

type Frame interface {
	DataRef() interface{}
	Dimensions() Dimensions
	Close()
}
