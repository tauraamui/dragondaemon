package dragon

import (
	"fmt"
	"sync"

	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

func (s *server) RunProcesses() {
	s.generateClipProcesses = map[string]process.Processable{}
	s.streamProcesses = map[string]process.Processable{}
	for _, cam := range s.cameras {
		frames := make(chan video.Frame)
		streamProcess := process.Settings{
			WaitForShutdownMsg: fmt.Sprintf("Closing camera [%s] video stream...", cam.Title()),
			Process:            process.StreamProcess(cam, frames),
		}
		s.streamProcesses[cam.Title()] = process.New(streamProcess)

		generateClipsFromFramesProcess := process.Settings{
			WaitForShutdownMsg: fmt.Sprintf("Stopping generating clips from [%s] video stream...", cam.Title()),
			Process:            process.GenerateClipsProcess(frames),
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
