package media

import (
	"github.com/tauraamui/dragondaemon/pkg/log"
	"gocv.io/x/gocv"
)

type videoClip struct {
	mockWriter bool
	fileName   string
	fps        int
	frames     []gocv.Mat
}

func (v *videoClip) flushToDisk() error {
	if len(v.frames) > 0 {
		img := v.frames[0]
		writer, err := openVideoWriter(
			v.fileName, "avc1.4d001e", float64(v.fps), img.Cols(), img.Rows(), v.mockWriter,
		)
		if err != nil {
			return err
		}

		defer writer.Close()

		log.Info("Saving to clip file: %s", v.fileName) //nolint

		for _, f := range v.frames {
			if f.Empty() {
				f.Close()
				continue
			}

			if writer.IsOpened() {
				if err := writer.Write(f); err != nil {
					log.Error("Unable to write frame to file: %v", err) //nolint
				}
			}
			f.Close()
		}
	}
	v.frames = nil
	return nil
}

func (v *videoClip) close() {
	for _, f := range v.frames {
		f.Close()
	}
}
