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
	"github.com/google/uuid"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/config"
	"github.com/tauraamui/dragondaemon/config/schedule"
	"gocv.io/x/gocv"
)

type Connection struct {
	inShutdown         int32
	attemptToReconnect chan bool
	uuid               string
	title              string
	persistLocation    string
	fps                int
	secondsPerClip     int
	sizeOnDisk         int64
	sizeOnDiskUnit     string
	schedule           schedule.Schedule
	reolinkControl     *reolinkapi.Camera
	mu                 sync.Mutex
	vc                 *gocv.VideoCapture
	rtspStream         string
	buffer             chan gocv.Mat
	window             *gocv.Window
}

func NewConnection(
	title string,
	persistLocation string,
	fps int,
	secondsPerClip int,
	schedule schedule.Schedule,
	reolink config.ReolinkAdvanced,
	vc *gocv.VideoCapture,
	rtspStream string,
) *Connection {
	var reolinkConn *reolinkapi.Camera
	if reolink.Enabled {
		conn, err := reolinkapi.NewCamera(reolink.Username, reolink.Password, reolink.APIAddress)
		if err != nil {
			logging.Error("Unable to get control connection for camera...")
		}

		reolinkConn = conn
	}

	return &Connection{
		attemptToReconnect: make(chan bool, 1),
		uuid:               uuid.NewString(),
		title:              title,
		persistLocation:    persistLocation,
		fps:                fps,
		secondsPerClip:     secondsPerClip,
		schedule:           schedule,
		reolinkControl:     reolinkConn,
		vc:                 vc,
		rtspStream:         rtspStream,
		buffer:             make(chan gocv.Mat, 6),
	}
}

func (c *Connection) ShowInWindow(winTitle string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.window != nil {
		c.window.Close()
	}

	c.window = gocv.NewWindow(winTitle)
	img := gocv.NewMat()
	defer img.Close()

	for atomic.LoadInt32(&c.inShutdown) == 0 {
		c.vc.Read(&img)
		if img.Empty() {
			continue
		}
		c.window.IMShow(img)
		c.window.WaitKey(1)
	}
}

func (c *Connection) UUID() string {
	return c.uuid
}

func (c *Connection) Title() string {
	return c.title
}

func (c *Connection) SizeOnDisk() (int64, string, error) {
	var size int64
	var unit string

	// TODO(tauraamui): will come up with some way to dirty the cache, maybe a timeout?
	if c.sizeOnDisk > 0 {
		size = c.sizeOnDisk
	}

	if len(c.sizeOnDiskUnit) > 0 {
		unit = c.sizeOnDiskUnit
	}

	if size > 0 && len(unit) > 0 {
		return size, unit, nil
	}

	startTime := time.Now()
	total, err := getDirSize(fmt.Sprintf("%s%c%s", c.persistLocation, os.PathSeparator, c.title), nil)
	endTime := time.Now()

	logging.Debug("FILE SIZE CHECK TOOK: %s", endTime.Sub(startTime))

	if err != nil {
		return total, "", err
	}

	c.sizeOnDisk, c.sizeOnDiskUnit = unitizeSize(total)

	return c.sizeOnDisk, c.sizeOnDiskUnit, nil
}

// will either get empty string and pointer or filled string with nil pointer
// depending on whether it needs to just count the files still remaining in
// this given dir or whether it needs to start counting again in a found sub dir
func getDirSize(path string, filePtr *os.File) (int64, error) {
	var total int64

	fp, err := resolveFilePointer(path, filePtr)
	if err != nil {
		return total, err
	}

	files, err := fp.Readdir(100)
	if len(files) == 0 {
		return total, err
	}

	for _, f := range files {
		if f.IsDir() {
			t, err := getDirSize(fmt.Sprintf("%s%c%s", path, os.PathSeparator, f.Name()), nil)
			if err == nil {
				total += t
			}
		}
		total += f.Size()
	}

	if err != io.EOF {
		t, err := getDirSize("", fp)
		if err == nil {
			total += t
		}
	}

	return total, nil
}

func resolveFilePointer(path string, file *os.File) (*os.File, error) {
	var fp *os.File
	fp = file
	if fp == nil {
		filePtr, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		fp = filePtr
	}
	return fp, nil
}

func unitizeSize(total int64) (int64, string) {
	unit := "Kb"
	total /= 1024
	if total > 1024 {
		total /= 1024
		unit = "Mb"
		if total > 1024 {
			total /= 1024
			unit = "Gb"
		}
	}

	return total, unit
}

func (c *Connection) Close() error {
	atomic.StoreInt32(&c.inShutdown, 1)
	if c.window != nil {
		c.window.Close()
	}
	close(c.buffer)
	return c.vc.Close()
}

func (c *Connection) persistToDisk() {
	img := <-c.buffer
	outputFile := fetchClipFilePath(c.persistLocation, c.title)
	writer, err := gocv.VideoWriterFile(outputFile, "avc1.4d001e", float64(c.fps), img.Cols(), img.Rows(), true)
	img.Close()

	if err != nil {
		logging.Error("Opening video writer device: %v", err)
	}
	defer writer.Close()

	logging.Info(fmt.Sprintf("Saving to clip file: %s", outputFile))

	var framesWritten uint
	for framesWritten = 0; framesWritten < uint(c.fps)*uint(c.secondsPerClip); framesWritten++ {
		img = <-c.buffer

		if img.Empty() {
			img.Close()
			return
		}

		if writer.IsOpened() {
			if err := writer.Write(img); err != nil {
				logging.Error("Unable to write frame to file: %v", err)
			}
		}
		img.Close()
	}
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
				// TODO(:tauraamui) Investigate why this case is reached more than once anyway
				if reachedShutdownCase == false {
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
					break
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
					defer imgClone.Close()
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

	c.vc, err = gocv.OpenVideoCapture(c.rtspStream)
	if err != nil {
		return err
	}

	return nil
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

func ensureDirectoryExists(path string) error {
	err := os.Mkdir(path, os.ModePerm)

	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}
