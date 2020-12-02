package media

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tacusci/logging"
	"gocv.io/x/gocv"
)

func fetchClipFilePath(rootDir string, clipsDir string) string {
	if len(rootDir) > 0 {
		ensureDirectoryExists(rootDir)
	} else {
		rootDir = "."
	}

	if len(clipsDir) > 0 {
		ensureDirectoryExists(fmt.Sprintf("%s/%s", rootDir, clipsDir))
	}

	return filepath.FromSlash(fmt.Sprintf("%s/%s/%s.avi", rootDir, clipsDir, time.Now().Format("2006-01-02 15.04.05")))
}

func ensureDirectoryExists(path string) error {
	err := os.Mkdir(path, os.ModePerm)

	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}

func (c *Connection) PersistToDisk(rootDir string, secondsPerClip uint) {
	img := gocv.NewMat()
	defer img.Close()

	c.vc.Read(&img)
	outputFile := fetchClipFilePath(rootDir, c.title)
	writer, err := gocv.VideoWriterFile(outputFile, "MJPG", 30, img.Cols(), img.Rows(), true)

	if err != nil {
		logging.Error(fmt.Sprintf("Opening video writer device: %v\n", err))
	}
	defer writer.Close()

	var framesWritten uint
	for framesWritten = 0; framesWritten < 30*secondsPerClip; framesWritten++ {
		if ok := c.vc.Read(&img); !ok {
			logging.Error(fmt.Sprintf("Device for stream at [%s] closed", c.title))
			return
		}
		if img.Empty() {
			logging.Debug("Skipping frame...")
			continue
		}

		if err := writer.Write(img); err != nil {
			logging.Error(fmt.Sprintf("Unable to write frame to file: %v", err))
		}
	}
}
