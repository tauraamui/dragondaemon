package videoframe

type Dimensions struct {
	W, H int
}

type Frame interface {
	NoCloser
	Closer
}

type NoCloser interface {
	DataRef() interface{}
	Dimensions() Dimensions
}

type Closer interface {
	Close()
}
