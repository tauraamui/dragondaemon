package media

import (
	"context"

	"github.com/ReolinkCameraAPI/reolinkapigo/pkg/reolinkapi"
	"github.com/spf13/afero"
	"gocv.io/x/gocv"
)

func OverloadLogConnDebug(overload func(
	format string, a ...interface{},
)) func() {
	logDebugRef := logConnDebug
	logConnDebug = overload
	return func() { logConnDebug = logDebugRef }
}

func OverloadLogConnInfo(overload func(
	format string, a ...interface{},
)) func() {
	logInfoRef := logConnInfo
	logConnInfo = overload
	return func() { logConnInfo = logInfoRef }
}

func OverloadLogConnWarn(overload func(
	format string, a ...interface{},
)) func() {
	logWarnRef := logConnWarn
	logConnWarn = overload
	return func() { logConnWarn = logWarnRef }
}

func OverloadLogConnError(overload func(
	format string, a ...interface{},
)) func() {
	logErrorRef := logConnError
	logConnError = overload
	return func() { logConnWarn = logErrorRef }
}

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

func OverloadOpenVideoWriter(overload func(
	string,
	string,
	float64,
	int,
	int,
) (VideoWriteable, error)) func() {
	openVideoWriterRef := openVideoWriter
	openVideoWriter = overload
	return func() { openVideoWriter = openVideoWriterRef }
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
