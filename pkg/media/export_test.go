package media

import (
	"context"
	"time"

	"github.com/ReolinkCameraAPI/reolinkapigo/pkg/reolinkapi"
	"github.com/allegro/bigcache/v3"
	"github.com/spf13/afero"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"gocv.io/x/gocv"
)

func OverloadNow(overload func() time.Time) func() {
	nowRef := now
	now = overload
	return func() { now = nowRef }
}

func OverloadLogInfo(overload func(string, ...interface{})) func() {
	logInfoRef := log.Info
	log.Info = overload
	return func() { log.Error = logInfoRef }
}

func OverloadLogError(overload func(string, ...interface{})) func() {
	logErrorRef := log.Error
	log.Error = overload
	return func() { log.Error = logErrorRef }
}

func OverloadFS(overload afero.Fs) func() {
	fsRef := fs
	fs = overload
	return func() { fs = fsRef }
}

func OverloadNewCache(overload func() (*bigcache.BigCache, error)) func() {
	newCacheRef := newCache
	newCache = overload
	return func() { newCache = newCacheRef }
}

func OverloadOpenVideoCapture(overload func(
	string,
	string,
	int,
	bool,
	string,
	bool,
) (VideoCapturable, error)) func() {
	openVidCapRef := openVideoCapture
	openVideoCapture = overload
	return func() {
		openVideoCapture = openVidCapRef
	}
}

func OverloadOpenVideoWriter(overload func(
	string,
	string,
	float64,
	int,
	int,
	bool,
) (VideoWriteable, error)) func() {
	openVideoWriterRef := openVideoWriter
	openVideoWriter = overload
	return func() { openVideoWriter = openVideoWriterRef }
}

func OverloadBeginProcesses(overload func(context.Context, Options, *Server) []processable) func() {
	beginProcessesRef := beginProcesses
	beginProcesses = overload
	return func() { beginProcesses = beginProcessesRef }
}

func OpenVideoWriter(
	fileName string,
	codec string,
	fps float64,
	frameWidth int,
	frameHeight int,
	mock bool,
) (VideoWriteable, error) {
	return openVideoWriter(fileName, codec, fps, frameWidth, frameHeight, mock)
}

func (c *Connection) Stream(ctx context.Context) chan interface{} {
	return c.stream(ctx)
}

func (c *Connection) WriteStreamToClips(ctx context.Context) chan interface{} {
	return c.writeStreamToClips(ctx)
}

func ReadFromStream(c *Connection, img *gocv.Mat) bool {
	return readFromStream(c, img)
}

func (c *Connection) Buffer() chan gocv.Mat {
	return c.buffer
}

func (c *Connection) Cache() *bigcache.BigCache {
	return c.cache
}

func (c *Connection) ReolinkControl() *reolinkapi.Camera {
	return c.reolinkControl
}

func UnitizeSize(total int64) (int64, string) {
	return unitizeSize(total)
}
