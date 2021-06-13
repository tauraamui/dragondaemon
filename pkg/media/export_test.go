package media

import (
	"context"

	"github.com/spf13/afero"
	"gocv.io/x/gocv"
)

func OverloadFS(overload afero.Fs) func() {
	fsRef := fs
	fs = overload
	return func() { fs = fsRef }
}

func (c *Connection) Stream(ctx context.Context) chan struct{} {
	return c.stream(ctx)
}

func (c *Connection) Buffer() chan gocv.Mat {
	return c.buffer
}
