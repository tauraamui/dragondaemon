package video

type Frame interface {
	DataRef() interface{}
	Dimensions() (int, int)
	Close()
}
