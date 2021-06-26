package media

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ReolinkCameraAPI/reolinkapigo/pkg/reolinkapi"
	"github.com/allegro/bigcache/v3"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/config"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"gocv.io/x/gocv"
)

const sizeOnDisk string = "sod"

var fs = afero.NewOsFs()

type ConnectonSettings struct {
	PersistLocation string
	FPS             int
	SecondsPerClip  int
	DateTimeLabel   bool
	DateTimeFormat  string
	Schedule        schedule.Schedule
	Reolink         config.ReolinkAdvanced
}

type Connection struct {
	sett               ConnectonSettings
	cache              *bigcache.BigCache
	inShutdown         int32
	attemptToReconnect chan bool
	uuid               string
	title              string
	reolinkControl     *reolinkapi.Camera
	mu                 sync.Mutex
	vc                 VideoCapturable
	rtspStream         string
	buffer             chan gocv.Mat
}

func NewConnection(
	title string,
	sett ConnectonSettings,
	vc VideoCapturable,
	rtspStream string,
) *Connection {
	control, err := connectReolinkControl(
		sett.Reolink.Username, sett.Reolink.Password, sett.Reolink.APIAddress,
	)
	if err != nil {
		logging.Error(err.Error())
	}

	cache, err := initCache()
	if err != nil {
		logging.Error(err.Error())
	}

	return &Connection{
		sett:               sett,
		cache:              cache,
		attemptToReconnect: make(chan bool, 1),
		uuid:               uuid.NewString(),
		title:              title,
		reolinkControl:     control,
		vc:                 vc,
		rtspStream:         rtspStream,
		buffer:             make(chan gocv.Mat, 6),
	}
}

func (c *Connection) UUID() string {
	return c.uuid
}

func (c *Connection) Title() string {
	return c.title
}

func (c *Connection) SizeOnDisk() (string, error) {
	var size int64
	var unit string
	var sizeWithUnitSuffix string

	s, err := c.cache.Get(sizeOnDisk)
	if err == nil {
		return string(s), nil
	} else {
		logging.Error("unable to retrieve size from cache: %w", err)
	}

	startTime := time.Now()
	total, err := getDirSize(filepath.Join(c.sett.PersistLocation, c.title), nil)
	endTime := time.Now()

	logging.Debug("FILE SIZE CHECK TOOK: %s", endTime.Sub(startTime))

	if err != nil {
		size, unit := unitizeSize(0)
		return fmt.Sprintf("%d%s", size, unit), err
	}

	size, unit = unitizeSize(total)
	sizeWithUnitSuffix = fmt.Sprintf("%d%s", size, unit)

	err = c.cache.Set(sizeOnDisk, []byte(sizeWithUnitSuffix))
	if err != nil {
		logging.Error("unable to store disk size in cache: %w", err)
	}

	return sizeWithUnitSuffix, nil
}

func (c *Connection) Close() error {
	atomic.StoreInt32(&c.inShutdown, 1)
	close(c.buffer)
	c.cache.Close()
	return c.vc.Close()
}

func (c *Connection) persistToDisk(ctx context.Context) chan interface{} {
	stopping := make(chan interface{})

	wg := sync.WaitGroup{}
	clipsToSave := make(chan videoClip, 3)
	go writeClipsToDisk(ctx, &wg, clipsToSave)

	go func(ctx context.Context, wg *sync.WaitGroup, stopping chan interface{}) {
		reachedShutdownCase := false
		for {
			time.Sleep(time.Millisecond * 1)
			select {
			case <-ctx.Done():
				if !reachedShutdownCase {
					reachedShutdownCase = true
					wg.Wait()
					close(clipsToSave)
					close(stopping)
				}
			default:
				clip := videoClip{
					fileName: fetchClipFilePath(c.sett.PersistLocation, c.title),
					frames:   []gocv.Mat{},
					fps:      c.sett.FPS,
				}
				// collect enough frames for clip
				var framesRead uint
				for framesRead = 0; framesRead < uint(c.sett.FPS*c.sett.SecondsPerClip); framesRead++ {
					clip.frames = append(clip.frames, <-c.buffer)
				}

				clipsToSave <- clip
			}
		}
	}(ctx, &wg, stopping)

	return stopping
}

func (c *Connection) stream(ctx context.Context) chan struct{} {
	logging.Debug("Opening root image mat")
	img := gocv.NewMat()

	stopping := make(chan struct{})

	reachedShutdownCase := false
	go func(ctx context.Context, stopping chan struct{}) {
		for {
			// throttle CPU usage
			time.Sleep(time.Millisecond * 1)
			select {
			case <-ctx.Done():
				if !reachedShutdownCase {
					reachedShutdownCase = true
					logging.Debug("Stopped stream goroutine")
					logging.Debug("Closing root image mat")
					img.Close()
					logging.Debug("Flushing img mat buffer")
					for len(c.buffer) > 0 {
						e := <-c.buffer
						e.Close()
					}
					close(stopping)
				}
			case reconnect := <-c.attemptToReconnect:
				if reconnect {
					logging.Info("Attempting to reconnect to [%s]", c.title)
					err := c.reconnect()
					if err != nil {
						logging.Error("Unable to reconnect to [%s]... ERROR: %v", c.title, err)
						c.attemptToReconnect <- true
						continue
					}
					logging.Info("Re-connected to [%s]...", c.title)
					continue
				}
			default:
				if c.vc.IsOpened() {
					if ok := c.vc.Read(&img); !ok {
						logging.Warn("Connection for stream at [%s] closed", c.title)
						c.attemptToReconnect <- true
						continue
					}

					imgClone := img.Clone()
					select {
					case c.buffer <- imgClone:
						logging.Debug("Sending read from to buffer...")
					default:
						imgClone.Close()
						logging.Debug("Buffer full...")
					}
				}
			}
		}
	}(ctx, stopping)

	return stopping
}

func (c *Connection) reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	if err = c.vc.Close(); err != nil {
		logging.Error("Failed to close connection... ERROR: %v", err)
	}

	vc, err := openVideoCapture(
		c.rtspStream,
		c.title,
		c.sett.FPS,
		c.sett.DateTimeLabel,
		c.sett.DateTimeFormat,
	)
	if err != nil {
		return err
	}

	c.vc = vc

	return nil
}

type VideoCapturable interface {
	SetP(*gocv.VideoCapture)
	IsOpened() bool
	Read(*gocv.Mat) bool
	Close() error
}

type videoWriteable interface {
	SetP(*gocv.VideoWriter)
	IsOpened() bool
	Write(gocv.Mat) error
	Close() error
}

func openVideoWriter(
	fileName string,
	codec string,
	fps float64,
	frameWidth int,
	frameHeight int,
) (videoWriteable, error) {
	mockVidWriting, foundEnv := os.LookupEnv("DRAGON_DAEMON_MOCK_VIDEO_WRITING")
	if foundEnv && mockVidWriting == "1" {
		return &mockVideoWriter{}, nil
	}
	vw, err := gocv.VideoWriterFile(fileName, codec, fps, frameWidth, frameHeight, true)
	if err != nil {
		return nil, err
	}

	return &videoWriter{
		p: vw,
	}, err
}

type videoWriter struct {
	p *gocv.VideoWriter
}

// SetP updates the internal pointer to the video capture instance.
func (vw *videoWriter) SetP(w *gocv.VideoWriter) {
	vw.p = w
}

func (vw *videoWriter) IsOpened() bool {
	return vw.p.IsOpened()
}

func (vw *videoWriter) Write(m gocv.Mat) error {
	return vw.p.Write(m)
}

func (vw *videoWriter) Close() error {
	return vw.p.Close()
}

type mockVideoWriter struct {
	initialised bool
}

// SetP doesn't do anything it exists to satisfy videoWriteable interface
func (mvw *mockVideoWriter) SetP(_ *gocv.VideoWriter) {}

// IsOpened always returns true.
func (mvw *mockVideoWriter) IsOpened() bool {
	return true
}

func (mvw *mockVideoWriter) Write(m gocv.Mat) error {
	return nil
}

func (mvw *mockVideoWriter) Close() error {
	mvw.initialised = false
	return nil
}

type videoClip struct {
	fileName string
	fps      int
	frames   []gocv.Mat
}

func (v *videoClip) writeToDisk() error {
	if len(v.frames) > 0 {
		img := v.frames[0]
		writer, err := openVideoWriter(v.fileName, "avc1.4d001e", float64(v.fps), img.Cols(), img.Rows())
		if err != nil {
			return err
		}

		defer writer.Close()

		logging.Info("Saving to clip file: %s", v.fileName)

		for _, f := range v.frames {
			if f.Empty() {
				f.Close()
				continue
			}

			if writer.IsOpened() {
				if err := writer.Write(f); err != nil {
					logging.Error("Unable to write frame to file: %v", err)
				}
			}
			f.Close()
		}
	}
	v.frames = nil
	return nil
}

func connectReolinkControl(username, password, addr string) (conn *reolinkapi.Camera, err error) {
	conn, err = reolinkapi.NewCamera(username, password, addr)
	if err != nil {
		err = fmt.Errorf("unable to connect to camera API: %w", err)
	}
	return
}

func initCache() (cache *bigcache.BigCache, err error) {
	cache, err = bigcache.NewBigCache(bigcache.DefaultConfig(5 * time.Minute))
	if err != nil {
		err = fmt.Errorf("unable to initialise connection cache: %w", err)
	}
	return
}

// will either get empty string and pointer or filled string with nil pointer
// depending on whether it needs to just count the files still remaining in
// this given dir or whether it needs to start counting again in a found sub dir
func getDirSize(path string, filePtr afero.File) (int64, error) {
	var total int64

	fp, err := resolveFilePointer(path, filePtr)
	if err != nil {
		return total, err
	}

	files, err := fp.Readdir(100)
	if len(files) == 0 {
		return total, err
	}

	total += countFileSizes(files, func(f os.FileInfo) int64 {
		done := make(chan interface{})

		var total int64
		go func(d chan interface{}, t *int64) {
			s, err := getDirSize(filepath.Join(path, f.Name()), nil)
			if err != nil {
				logging.Error("Unable to get dirs full size: %v...", err)
			}
			*t += s
			close(done)
		}(done, &total)

		<-done
		return total
	})

	if err != io.EOF {
		t, err := getDirSize("", fp)
		if err == nil {
			total += t
		}
	}

	return total, nil
}

func fetchClipFilePath(rootDir string, clipsDir string) string {
	if len(rootDir) > 0 {
		err := ensureDirectoryExists(rootDir)
		if err != nil {
			logging.Error("Unable to create directory %s: %v", rootDir, err)
		}
	} else {
		rootDir = "."
	}

	todaysDate := time.Now().Format("2006-01-02")

	if len(clipsDir) > 0 {
		path := fmt.Sprintf("%s/%s", rootDir, clipsDir)
		err := ensureDirectoryExists(path)
		if err != nil {
			logging.Error("Unable to create directory %s: %v", path, err)
		}

		path = fmt.Sprintf("%s/%s/%s", rootDir, clipsDir, todaysDate)
		err = ensureDirectoryExists(path)
		if err != nil {
			logging.Error("Unable to create directory %s: %v", path, err)
		}
	}

	return filepath.FromSlash(fmt.Sprintf("%s/%s/%s/%s.mp4", rootDir, clipsDir, todaysDate, time.Now().Format("2006-01-02 15.04.05")))
}

func writeClipsToDisk(ctx context.Context, wg *sync.WaitGroup, clips chan videoClip) {
	readAndWrite := func(clips chan videoClip) {
		clip := <-clips
		if err := clip.writeToDisk(); err != nil {
			logging.Error("Unable to write video clip %s to disk: %v", clip.fileName, err)
		}
	}
	wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup, clips chan videoClip) {
		defer wg.Done()
		for {
			time.Sleep(time.Millisecond * 1)
			select {
			case <-ctx.Done():
				for len(clips) > 0 {
					readAndWrite(clips)
				}
				return
			default:
				readAndWrite(clips)
			}
		}
	}(ctx, wg, clips)
}

func ensureDirectoryExists(path string) error {
	err := os.Mkdir(path, os.ModePerm)

	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}

func countFileSizes(files []os.FileInfo, onDirFile func(os.FileInfo) int64) int64 {
	var total int64
	for _, f := range files {
		if f.IsDir() {
			total += onDirFile(f)
			continue
		}
		if f.Mode().IsRegular() {
			total += f.Size()
		}
	}
	return total
}

func resolveFilePointer(path string, file afero.File) (afero.File, error) {
	var fp afero.File
	fp = file
	if fp == nil {
		filePtr, err := fs.Open(path)
		if err != nil {
			return nil, err
		}
		fp = filePtr
	}
	return fp, nil
}

func unitizeSize(total int64) (int64, string) {
	unit := "KB"
	total /= 1024
	if total >= 1024 {
		total /= 1024
		unit = "MB"
		if total >= 1024 {
			total /= 1024
			unit = "GB"
		}
	}

	return total, unit
}
