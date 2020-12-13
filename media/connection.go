package media

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tacusci/logging"
	"gocv.io/x/gocv"
)

type Connection struct {
	inShutdown      int32
	title           string
	persistLocation string
	secondsPerClip  int
	vc              *gocv.VideoCapture
	mu              sync.Mutex
	window          *gocv.Window
}

func NewConnection(
	title string,
	persistLocation string,
	secondsPerClip int,
	vc *gocv.VideoCapture,
) *Connection {
	return &Connection{
		title:           title,
		persistLocation: persistLocation,
		secondsPerClip:  secondsPerClip,
		vc:              vc,
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

func (c *Connection) Title() string {
	return c.title
}

func (c *Connection) PersistToDisk() {
	img := c.FetchFrame()
	defer img.Close()
	outputFile := fetchClipFilePath(c.persistLocation, c.title)
	writer, err := gocv.VideoWriterFile(outputFile, "mp4v", 30, img.Cols(), img.Rows(), true)

	if err != nil {
		logging.Error(fmt.Sprintf("Opening video writer device: %v\n", err))
	}
	defer writer.Close()

	var framesWritten uint
	for framesWritten = 0; framesWritten < 30*uint(c.secondsPerClip); framesWritten++ {
		if ok := c.vc.Read(img); !ok {
			logging.Error(fmt.Sprintf("Device for stream at [%s] closed", c.title))
			return
		}

		if img.Empty() {
			logging.Debug("Skipping frame...")
			continue
		}

		if err := writer.Write(*img); err != nil {
			logging.Error(fmt.Sprintf("Unable to write frame to file: %v", err))
		}
	}
}

func (c *Connection) FetchFrame() *gocv.Mat {
	img := gocv.NewMat()
	c.vc.Read(&img)
	return &img
}

func (c *Connection) Close() error {
	atomic.StoreInt32(&c.inShutdown, 1)
	if c.window != nil {
		c.window.Close()
	}
	return c.vc.Close()
}

func fetchClipFilePath(rootDir string, clipsDir string) string {
	if len(rootDir) > 0 {
		ensureDirectoryExists(rootDir)
	} else {
		rootDir = "."
	}

	todaysDate := time.Now().Format("2006-01-02")

	if len(clipsDir) > 0 {
		ensureDirectoryExists(fmt.Sprintf("%s/%s/%s", rootDir, clipsDir, todaysDate))
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
