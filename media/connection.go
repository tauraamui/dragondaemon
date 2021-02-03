package media

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tacusci/logging/v2"
	"gocv.io/x/gocv"
)

type Connection struct {
	inShutdown         int32
	attemptToReconnect chan bool
	title              string
	persistLocation    string
	fps                int
	secondsPerClip     int
	mu                 sync.Mutex
	vc                 *gocv.VideoCapture
	rtspStream         string
	lastFrameData      gocv.Mat
	buffer             chan *gocv.Mat
	window             *gocv.Window
}

func NewConnection(
	title string,
	persistLocation string,
	fps int,
	secondsPerClip int,
	vc *gocv.VideoCapture,
	rtspStream string,
) *Connection {
	return &Connection{
		attemptToReconnect: make(chan bool, 1),
		title:              title,
		persistLocation:    persistLocation,
		fps:                fps,
		secondsPerClip:     secondsPerClip,
		vc:                 vc,
		rtspStream:         rtspStream,
		lastFrameData:      gocv.NewMat(),
		buffer:             make(chan *gocv.Mat, 6),
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
	defer img.Close()
	outputFile := fetchClipFilePath(c.persistLocation, c.title)
	writer, err := gocv.VideoWriterFile(outputFile, "avc1.4d001e", float64(c.fps), img.Cols(), img.Rows(), true)

	if err != nil {
		logging.Error("Opening video writer device: %v", err)
	}
	defer writer.Close()

	logging.Info(fmt.Sprintf("Saving to clip file: %s", outputFile))

	var framesWritten uint
	for framesWritten = 0; framesWritten < uint(c.fps)*uint(c.secondsPerClip); framesWritten++ {
		img = <-c.buffer
		defer img.Close()

		if img.Empty() {
			continue
		}

		if writer.IsOpened() {
			if err := writer.Write(*img); err != nil {
				logging.Error("Unable to write frame to file: %v", err)
			}
		}
	}
}

func (c *Connection) stream(stop chan struct{}) {
	for {
		// throttle CPU usage
		time.Sleep(time.Millisecond * 100)
		select {
		case <-stop:
			c.lastFrameData.Close()
			break
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
				img := gocv.NewMat()

				if ok := c.vc.Read(&img); !ok {
					logging.Warn("Connection for stream at [%s] closed", c.title)
					c.attemptToReconnect <- true
					continue
				}

				img.CopyTo(&c.lastFrameData)
				bufferClone := c.lastFrameData.Clone()
				select {
				case c.buffer <- &bufferClone:
				default:
				}

				img.Close()
			}
		}
	}
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
