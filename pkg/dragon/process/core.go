package process

import (
	"errors"

	"github.com/spf13/afero"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

var fs afero.Fs = afero.NewOsFs()

func NewCoreProcess(cam camera.Connection, writer video.ClipWriter) Process {
	return &persistCameraToDisk{
		cam:    cam,
		writer: writer,
		frames: make(chan video.Frame, 3),
		clips:  make(chan video.Clip, 3),
	}
}

type persistCameraToDisk struct {
	cam           camera.Connection
	writer        video.ClipWriter
	frames        chan video.Frame
	clips         chan video.Clip
	streamProcess Process
	generateClips Process
	persistClips  Process
}

func (proc *persistCameraToDisk) Setup() {
	proc.streamProcess = NewStreamConnProcess(proc.cam, proc.frames)
	proc.streamProcess.RegisterCallback(PROC_CAM_SWITCHED_OFF, func() {})
	proc.generateClips = NewGenerateClipProcess(
		proc.frames, proc.clips, proc.cam.FPS()*proc.cam.SPC(), proc.cam.FullPersistLocation(),
	)
	proc.persistClips = NewPersistClipProcess(proc.clips, proc.writer)
}

func (proc *persistCameraToDisk) RegisterCallback(code event, callback func()) error {
	return errors.New("persist camera proc does not support event callbacks")
}

func (proc *persistCameraToDisk) Start() {
	log.Info("Streaming video from camera [%s]", proc.cam.Title())
	proc.streamProcess.Start()
	log.Info("Generating clips from camera [%s] video stream...", proc.cam.Title())
	proc.generateClips.Start()
	log.Info("Writing clips to disk from camera [%s] video stream...", proc.cam.Title())
	proc.persistClips.Start()
}

func (proc *persistCameraToDisk) Stop() {
	log.Info("Stopping writing clips to disk from camera [%s] video stream...", proc.cam.Title())
	proc.persistClips.Stop()
	log.Info("Stopping generating clips from camera [%s] video stream...", proc.cam.Title())
	proc.generateClips.Stop()
	log.Info("Closing camera [%s] video stream...", proc.cam.Title())
	proc.streamProcess.Stop()
}

func (proc *persistCameraToDisk) Wait() {
	log.Info("Waiting for writing clips to disk shutdown...")
	proc.persistClips.Wait()
	log.Info("Waiting for generating clips to shutdown...")
	proc.generateClips.Wait()
	log.Info("Waiting for streaming video to shutdown...")
	proc.streamProcess.Wait()
}
