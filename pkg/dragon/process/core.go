package process

import (
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
	proc.generateClips = NewGenerateClipProcess(proc.frames, proc.clips, proc.cam.FPS()*proc.cam.SPC(), proc.cam.FullPersistLocation())
	proc.persistClips = NewPersistClipProcess(proc.clips, proc.writer)
	// writeClipsToDiskProcess := Settings{
	// 	WaitForShutdownMsg: fmt.Sprintf("Stopping writing clips to disk from [%s] video stream...", proc.cam.Title()),
	// 	Process:            WriteClipsToDiskProcess(proc.clips, proc.writer),
	// }
	// proc.writeClips = New(writeClipsToDiskProcess)

	// generateClipsFromFramesProcess := Settings{
	// 	WaitForShutdownMsg: fmt.Sprintf("Stopping generating clips from [%s] video stream...", proc.cam.Title()),
	// 	Process: GenerateClipsProcess(
	// 		proc.frames, proc.clips, proc.cam.FullPersistLocation(), proc.cam.FPS(), proc.cam.SPC(),
	// 	),
	// }
	// proc.generateClips = New(generateClipsFromFramesProcess)

	// streamProcess := Settings{
	// 	WaitForShutdownMsg: fmt.Sprintf("Closing camera [%s] video stream...", proc.cam.Title()),
	// 	Process:            StreamProcess(proc.cam, proc.frames),
	// }
	// proc.streamProcess = New(streamProcess)

	// deleteProcess := Settings{
	// 	WaitForShutdownMsg: fmt.Sprintf("Stopping deleting old saved clips for [%s]", proc.cam.Title()),
	// 	Process:            DeleteOldClips(proc.cam),
	// }
	// proc.deleteClips = New(deleteProcess)
}

func (proc *persistCameraToDisk) Start() {
	log.Info("Streaming video from camera [%s]", proc.cam.Title())
	proc.streamProcess.Start()
	log.Info("Generating clips from camera [%s] video stream...", proc.cam.Title())
	proc.generateClips.Start()
	log.Info("Writing clips to disk from camera [%s] video stream...", proc.cam.Title())
	proc.persistClips.Start()
	// proc.deleteClips.Start()
	// proc.writeClips.Start()
	// proc.streamProcess.Start()
	// go func(clips chan video.Clip) {
	// 	for clip := range clips {
	// 		log.Info("Closing clip from camera [%s]", proc.cam.Title())
	// 		clip.Close()
	// 	}
	// }(proc.clips)
}

func (proc *persistCameraToDisk) Stop() {
	// proc.deleteClips.Stop()
	// proc.writeClips.Stop()
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
