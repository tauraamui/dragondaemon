package main

import (
	"fmt"
	"os"

	"gocv.io/x/gocv"
)

// "rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov"

func main() {
	saveFile := os.Args[0]

	webcam, err := gocv.OpenVideoCapture("rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mov")
	if err != nil {
		fmt.Println("Error opening video capture device")
		return
	}
	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	if ok := webcam.Read(&img); !ok {
		fmt.Println("Unable to read from IP camera")
		return
	}

	writer, err := gocv.VideoWriterFile(saveFile, "MJPG", 25, img.Cols(), img.Rows(), true)
	if err != nil {
		fmt.Printf("error opening video writer device: %v\n", err)
		return
	}
	defer writer.Close()

	for i := 0; i < 100; i++ {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed\n")
			return
		}
		if img.Empty() {
			continue
		}

		writer.Write(img)
	}
}

func writeFrameToFile(writer gocv.VideoWriter, frame gocv.Mat) error {
	return nil
}
