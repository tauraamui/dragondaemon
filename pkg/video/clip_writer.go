package video

type ClipWriter interface {
	Write(ClipNoCloser) error
}
