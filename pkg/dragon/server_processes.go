package dragon

import (
	"sync"

	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
)

func (s *server) SetupProcesses() {
	for _, cam := range s.cameras {
		proc := process.NewCoreProcess(cam, s.videoBackend.NewWriter())
		proc.Setup()
		s.coreProcesses = append(s.coreProcesses, proc)
	}
}

func (s *server) RunProcesses() {
	for _, proc := range s.coreProcesses {
		proc.Start()
	}
}

func (s *server) shutdownProcesses() {
	wg := sync.WaitGroup{}
	wg.Add(len(s.coreProcesses))
	for _, proc := range s.coreProcesses {
		go func(wg *sync.WaitGroup, proc process.Process) {
			proc.Stop()
			proc.Wait()
			wg.Done()
		}(&wg, proc)
	}
	wg.Wait()
}
