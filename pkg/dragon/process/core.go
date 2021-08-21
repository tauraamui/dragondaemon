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
	writeClips    Process
	deleteClips   Process
}

func (proc *persistCameraToDisk) Setup() {
	writeClipsToDiskProcess := Settings{
		WaitForShutdownMsg: fmt.Sprintf("Stopping writing clips to disk from [%s] video stream...", proc.cam.Title()),
		Process:            WriteClipsToDiskProcess(proc.clips),
	}
	proc.writeClips = New(writeClipsToDiskProcess)

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

	deleteProcess := Settings{
		WaitForShutdownMsg: fmt.Sprintf("Stopping deleting old saved clips for [%s]", proc.cam.Title()),
		Process:            DeleteOldClips(proc.cam),
	}
	proc.deleteClips = New(deleteProcess)
}

func (proc *persistCameraToDisk) Start() {
	proc.deleteClips.Start()
	proc.writeClips.Start()
	proc.generateClips.Start()
	proc.streamProcess.Start()
}

func (proc *persistCameraToDisk) Stop() {
	proc.deleteClips.Stop()
	proc.writeClips.Stop()
	proc.generateClips.Stop()
	proc.streamProcess.Stop()
}

func (proc *persistCameraToDisk) Wait() {
	wg := sync.WaitGroup{}
	wg.Add(4)
	go func(wg *sync.WaitGroup) {
		proc.deleteClips.Wait()
		wg.Done()
	}(&wg)
	go func(wg *sync.WaitGroup) {
		proc.writeClips.Wait()
		wg.Done()
	}(&wg)
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
