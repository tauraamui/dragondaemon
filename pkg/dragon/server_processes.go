package dragon

import (
	"context"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

func streamProcess(s *server, frames chan video.Frame) func(cancel context.Context) []chan interface{} {
	return func(cancel context.Context) []chan interface{} {
		var stopSignals []chan interface{}
		for _, cam := range s.cameras {
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
			}(cancel, cam, stopping)
			stopSignals = append(stopSignals, stopping)
		}
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
					log.Debug("Reading frame from channel")
					f := <-frames
					f.Close()
				}
			}
		}(frames, stopping)
		stopSignals = append(stopSignals, stopping)
		return stopSignals
	}
}

func (s *server) RunProcesses() {
	frames := make(chan video.Frame)

	streamProcess := process.Settings{
		WaitForShutdownMsg: "Stopping stream process",
		Process:            streamProcess(s, frames),
	}

	generateClipsFromFramesProcess := process.Settings{
		WaitForShutdownMsg: "Stopping building clips from vid stream",
		Process:            generateClipsProcess(frames),
	}

	s.processes = append(s.processes, process.New(generateClipsFromFramesProcess))
	s.processes = append(s.processes, process.New(streamProcess))

	for _, proc := range s.processes {
		proc.Start()
	}
}

func (s *server) shutdownProcesses() {
	for _, proc := range s.processes {
		proc.Stop()
		proc.Wait()
	}
}
