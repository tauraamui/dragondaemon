package process

import (
	"fmt"
	"sync"

	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

func NewCoreProcess(cam camera.Connection) Process {
	return &persistCameraToDisk{
		cam:    cam,
		frames: make(chan video.Frame),
		clips:  make(chan video.Clip),
	}
}

type persistCameraToDisk struct {
	cam           camera.Connection
	frames        chan video.Frame
	clips         chan video.Clip
	streamProcess Process
	generateClips Process
}

func (proc *persistCameraToDisk) Setup() {
	generateClipsFromFramesProcess := Settings{
		WaitForShutdownMsg: fmt.Sprintf("Stopping generating clips from [%s] video stream...", proc.cam.Title()),
		Process:            GenerateClipsProcess(proc.frames, proc.clips, proc.cam.FPS(), proc.cam.SPC()),
	}
	proc.generateClips = New(generateClipsFromFramesProcess)

	streamProcess := Settings{
		WaitForShutdownMsg: fmt.Sprintf("Closing camera [%s] video stream...", proc.cam.Title()),
		Process:            StreamProcess(proc.cam, proc.frames),
	}
	proc.streamProcess = New(streamProcess)
}

func (proc *persistCameraToDisk) Start() {
	proc.generateClips.Start()
	proc.streamProcess.Start()
}

func (proc *persistCameraToDisk) Stop() {
	proc.generateClips.Stop()
	proc.streamProcess.Stop()
}

func (proc *persistCameraToDisk) Wait() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		proc.generateClips.Wait()
		wg.Done()
	}(&wg)
	go func(wg *sync.WaitGroup) {
		proc.streamProcess.Wait()
		wg.Done()
	}(&wg)
	wg.Wait()
}
