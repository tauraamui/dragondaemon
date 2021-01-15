package media

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"gocv.io/x/gocv"
)

type Connection struct {
	stdlog, errlog  *log.Logger
	inShutdown      int32
	title           string
	persistLocation string
	secondsPerClip  int
	mu              sync.Mutex
	vc              *gocv.VideoCapture
	lastFrameData   gocv.Mat
	buffer          chan *gocv.Mat
	window          *gocv.Window
}

func NewConnection(
	stdlog, errlog *log.Logger,
	title string,
	persistLocation string,
	secondsPerClip int,
	vc *gocv.VideoCapture,
) *Connection {
	return &Connection{
		stdlog:          stdlog,
		errlog:          errlog,
		title:           title,
		persistLocation: persistLocation,
		secondsPerClip:  secondsPerClip,
		vc:              vc,
		lastFrameData:   gocv.NewMat(),
		buffer:          make(chan *gocv.Mat, 3),
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
	outputFile := fetchClipFilePath(c.errlog, c.persistLocation, c.title)
	writer, err := gocv.VideoWriterFile(outputFile, "avc1.4d001e", 30, img.Cols(), img.Rows(), true)

	if err != nil {
		c.errlog.Printf("Opening video writer device: %v\n", err)
	}
	defer writer.Close()

	c.stdlog.Printf("Saving to clip file: %s\n", outputFile)

	var framesWritten uint
	for framesWritten = 0; framesWritten < 30*uint(c.secondsPerClip); framesWritten++ {
		img = <-c.buffer
		defer img.Close()

		if img.Empty() {
			continue
		}

		if writer.IsOpened() {
			if err := writer.Write(*img); err != nil {
				c.errlog.Printf("Unable to write frame to file: %v\n", err)
			}
		}
	}
}

func (c *Connection) stream(stop chan struct{}) {
	for {
		// throttle CPU usage
		time.Sleep(time.Millisecond * 1)
		select {
		case <-stop:
			c.lastFrameData.Close()
			break
		default:
			if c.vc.IsOpened() {
				img := gocv.NewMat()

				if ok := c.vc.Read(&img); !ok {
					c.stdlog.Printf("WARN: Device for stream at [%s] closed\n", c.title)
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

func fetchClipFilePath(errlog *log.Logger, rootDir string, clipsDir string) string {
	if len(rootDir) > 0 {
		err := ensureDirectoryExists(rootDir)
		if err != nil {
			errlog.Printf("Unable to create directory %s: %v\n", rootDir, err)
		}
	} else {
		rootDir = "."
	}

	todaysDate := time.Now().Format("2006-01-02")

	if len(clipsDir) > 0 {
		path := fmt.Sprintf("%s/%s", rootDir, clipsDir)
		err := ensureDirectoryExists(path)
		if err != nil {
			errlog.Printf("Unable to create directory %s: %v\n", path, err)
		}

		path = fmt.Sprintf("%s/%s/%s", rootDir, clipsDir, todaysDate)
		err = ensureDirectoryExists(path)
		if err != nil {
			errlog.Printf("Unable to create directory %s: %v\n", path, err)
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
