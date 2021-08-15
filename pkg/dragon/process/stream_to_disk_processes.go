package process

import (
	"context"
	"fmt"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

func StreamProcess(cam camera.Connection, frames chan video.Frame) func(cancel context.Context) []chan interface{} {
	return func(cancel context.Context) []chan interface{} {
		var stopSignals []chan interface{}
		log.Info("Streaming video from camera [%s]", cam.Title())
		stopping := make(chan interface{})
		go func(cancel context.Context, cam camera.Connection, stopping chan interface{}) {
		procLoop:
			for {
				time.Sleep(1 * time.Microsecond)
				select {
				case <-cancel.Done():
					close(stopping)
					break procLoop
				default:
					stream(cam, frames)
				}
			}
		}(cancel, cam, stopping)
		stopSignals = append(stopSignals, stopping)
		return stopSignals
	}
}

var stream = func(cam camera.Connection, frames chan video.Frame) {
	if cam.IsOpen() {
		log.Debug("Reading frame from vid stream for camera [%s]", cam.Title())
		frame := cam.Read()
		select {
		case frames <- frame:
			log.Debug("Sending frame from cam to buffer...")
		default:
			frame.Close()
			log.Debug("Buffer full...")
		}
	}
}

func GenerateClipsProcess(frames chan video.Frame, clips chan video.Clip, fps int, spc int) func(cancel context.Context) []chan interface{} {
	return func(cancel context.Context) []chan interface{} {
		var stopSignals []chan interface{}
		stopping := make(chan interface{})
		go func(frames chan video.Frame, stopping chan interface{}) {
		procLoop:
			for {
				time.Sleep(1 * time.Microsecond)
				select {
				case <-cancel.Done():
					close(stopping)
					break procLoop
				default:
					clips <- generateClipFromStream(frames, fps, spc)
				}
			}
		}(frames, stopping)
		stopSignals = append(stopSignals, stopping)
		return stopSignals
	}
}

func generateClipFromStream(frames chan video.Frame, fps, spc int) video.Clip {
	clip := video.NewClip()
	for framesRead := 0; framesRead < fps*spc; framesRead++ {
		frame := <-frames
		clip.AppendFrame(frame)
	}
	return clip
}

func WriteClipsToDiskProcess(clips chan video.Clip) func(canel context.Context) []chan interface{} {
	return func(cancel context.Context) []chan interface{} {
		var stopSignals []chan interface{}
		stopping := make(chan interface{})
		go func(clips chan video.Clip, stopping chan interface{}) {
		procLoop:
			for {
				time.Sleep(1 * time.Microsecond)
				select {
				case <-cancel.Done():
					close(stopping)
					break procLoop
				default:
					writeClipToDisk(<-clips)
				}
			}
		}(clips, stopping)
		stopSignals = append(stopSignals, stopping)
		return stopSignals
	}
}

var writeClipToDisk = func(clip video.Clip) {
	err := clip.Write()
	if err != nil {
		log.Error(fmt.Errorf("Unable to write clip to disk: %w", err).Error())
		clip.Close()
	}
}
