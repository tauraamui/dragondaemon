package media

import (
	"context"

	"github.com/ReolinkCameraAPI/reolinkapigo/pkg/reolinkapi"
	"github.com/spf13/afero"
	"gocv.io/x/gocv"
)

func OverloadFS(overload afero.Fs) func() {
	fsRef := fs
	fs = overload
	return func() { fs = fsRef }
}

func OverloadOpenVideoCapture(overload func(
	string,
	string,
	int,
	bool,
	string,
) (VideoCapturable, error)) func() {
	openVidCapRef := openVideoCapture
	openVideoCapture = overload
	return func() { openVideoCapture = openVidCapRef }
}

func (c *Connection) Stream(ctx context.Context) chan struct{} {
	return c.stream(ctx)
}

func (c *Connection) WriteStreamToClips(ctx context.Context) chan interface{} {
	return c.writeStreamToClips(ctx)
}

func (c *Connection) Buffer() chan gocv.Mat {
	return c.buffer
}

func (c *Connection) ReolinkControl() *reolinkapi.Camera {
	return c.reolinkControl
}

func UnitizeSize(total int64) (int64, string) {
	return unitizeSize(total)
}
