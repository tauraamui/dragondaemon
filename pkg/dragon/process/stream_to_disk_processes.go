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

func stream(cam camera.Connection, frames chan video.Frame) {
	if cam.IsOpen() {
		log.Debug("Reading frame from vid stream for camera [%s]", cam.Title())
		frame, err := cam.Read()
		if err != nil {
			log.Error(fmt.Errorf("Unable to retrieve frame: %w. Auto re-connecting is not yet implemented", err).Error())
			return
		}
		select {
		case frames <- frame:
			log.Debug("Sending frame from cam to buffer...")
		default:
			frame.Close()
			log.Debug("Buffer full...")
		}
	}
}

func GenerateClipsProcess(
	frames chan video.Frame,
	clips chan video.Clip,
	fullCamPersistLocation string,
	fps, spc int,
) func(cancel context.Context) []chan interface{} {
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
					clips <- generateClipFromStream(cancel, frames, fullCamPersistLocation, fps, spc)
				}
			}
		}(frames, stopping)
		stopSignals = append(stopSignals, stopping)
		return stopSignals
	}
}

func generateClipFromStream(cancel context.Context, frames chan video.Frame, persistLocation string, fps, spc int) video.Clip {
	clip := video.NewClip(persistLocation, fps)

	var capturedFrames int
procLoop:
	for capturedFrames < fps*spc {
		select {
		case <-cancel.Done():
			break procLoop
		default:
			clip.AppendFrame(<-frames)
			capturedFrames++
		}
	}
	return clip
}

func WriteClipsToDiskProcess(clips chan video.Clip, writer video.ClipWriter) func(canel context.Context) []chan interface{} {
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
					writeClipToDisk(<-clips, writer)
				}
			}
		}(clips, stopping)
		stopSignals = append(stopSignals, stopping)
		return stopSignals
	}
}

func writeClipToDisk(clip video.Clip, writer video.ClipWriter) {
	err := writer.Write(clip)
	if err != nil {
		log.Error(fmt.Errorf("Unable to write clip to disk: %w", err).Error())
	}
	clip.Close()
}
