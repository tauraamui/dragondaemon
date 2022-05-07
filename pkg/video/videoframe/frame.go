package videoframe

type Dimensions struct {
	W, H int
}

type Frame interface {
	NoCloser
	Closer
}

type NoCloser interface {
	Timestamp() int64
	Dimensions() Dimensions
	DataRef() interface{}
	ToBytes() []byte
}

type Closer interface {
	Close()
}
