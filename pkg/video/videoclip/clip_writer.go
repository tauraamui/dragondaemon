package videoclip

type Writer interface {
	Write(NoCloser) error
}
