package video

type Frame interface {
	DataRef() interface{}
	Close()
}
