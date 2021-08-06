package dragon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

func streamProcess(cam camera.Connection, frames chan video.Frame) func(cancel context.Context) []chan interface{} {
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
			}
		}(cancel, cam, stopping)
		stopSignals = append(stopSignals, stopping)
		return stopSignals
	}
}

func generateClipsProcess(frames chan video.Frame) func(cancel context.Context) []chan interface{} {
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
					select {
					case f := <-frames:
						log.Debug("Reading frame from channel")
						f.Close()
					default:
					}
				}
			}
		}(frames, stopping)
		stopSignals = append(stopSignals, stopping)
		return stopSignals
	}
}

func (s *server) RunProcesses() {
	s.generateClipProcesses = map[string]process.Processable{}
	s.streamProcesses = map[string]process.Processable{}
	for _, cam := range s.cameras {
		frames := make(chan video.Frame)
		streamProcess := process.Settings{
			WaitForShutdownMsg: fmt.Sprintf("Closing camera [%s] video stream...", cam.Title()),
			Process:            streamProcess(cam, frames),
		}
		s.streamProcesses[cam.Title()] = process.New(streamProcess)

		generateClipsFromFramesProcess := process.Settings{
			WaitForShutdownMsg: fmt.Sprintf("Stopping generating clips from [%s] video stream...", cam.Title()),
			Process:            generateClipsProcess(frames),
		}
		s.generateClipProcesses[cam.Title()] = process.New(generateClipsFromFramesProcess)
	}

	for _, proc := range s.generateClipProcesses {
		proc.Start()
	}

	for _, proc := range s.streamProcesses {
		proc.Start()
	}
}

func (s *server) shutdownProcesses() {
	wg := sync.WaitGroup{}
	wg.Add(len(s.generateClipProcesses))
	wg.Add(len(s.streamProcesses))

	for _, proc := range s.generateClipProcesses {
		proc.Stop()
		go func(wg *sync.WaitGroup, proc process.Processable) {
			defer wg.Done()
			proc.Wait()
		}(&wg, proc)
	}

	for _, proc := range s.streamProcesses {
		proc.Stop()
		go func(wg *sync.WaitGroup, proc process.Processable) {
			defer wg.Done()
			proc.Wait()
		}(&wg, proc)
	}
}
