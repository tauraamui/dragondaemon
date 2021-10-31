package video

type ClipWriter interface {
	Write(Clip) error
}
