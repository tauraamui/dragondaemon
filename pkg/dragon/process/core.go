package process

import (
	"context"
	"time"

	"github.com/spf13/afero"
	"github.com/tauraamui/dragondaemon/pkg/broadcast"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/dragondaemon/pkg/video"
)

var fs afero.Fs = afero.NewOsFs()

func NewCoreProcess(cam camera.Connection, writer video.ClipWriter) Process {
	return &persistCameraToDisk{
		broadcaster: broadcast.New(0),
		cam:         cam,
		writer:      writer,
		frames:      make(chan video.Frame, 3),
		clips:       make(chan video.Clip, 3),
	}
}

type persistCameraToDisk struct {
	broadcaster          *broadcast.Broadcaster
	cam                  camera.Connection
	writer               video.ClipWriter
	frames               chan video.Frame
	clips                chan video.Clip
	monitorCameraOnState Process
	streamProcess        Process
	generateClips        Process
	persistClips         Process
}

func (proc *persistCameraToDisk) Setup() Process {
	proc.monitorCameraOnState = New(Settings{
		WaitForShutdownMsg: "",
		Process:            sendEvtOnCameraStateChange(proc.broadcaster, proc.cam),
	})
	proc.streamProcess = NewStreamConnProcess(proc.broadcaster.Listen(), proc.cam, proc.frames)
	proc.generateClips = NewGenerateClipProcess(
		proc.broadcaster.Listen(), proc.frames, proc.clips, proc.cam.FPS()*proc.cam.SPC(), proc.cam.FullPersistLocation(),
	)
	proc.persistClips = NewPersistClipProcess(proc.clips, proc.writer)
	return proc
}

func (proc *persistCameraToDisk) Start() {
	log.Debug("Monitoring camera on/off state change")
	proc.monitorCameraOnState.Start()
	log.Info("Streaming video from camera [%s]", proc.cam.Title())
	proc.streamProcess.Start()
	log.Info("Generating clips from camera [%s] video stream...", proc.cam.Title())
	proc.generateClips.Start()
	log.Info("Writing clips to disk from camera [%s] video stream...", proc.cam.Title())
	proc.persistClips.Start()
}

func (proc *persistCameraToDisk) Stop() {
	log.Debug("Stopping monitoring camera on/off state change")
	proc.monitorCameraOnState.Stop()
	log.Info("Stopping writing clips to disk from camera [%s] video stream...", proc.cam.Title())
	proc.persistClips.Stop()
	log.Info("Stopping generating clips from camera [%s] video stream...", proc.cam.Title())
	proc.generateClips.Stop()
	log.Info("Closing camera [%s] video stream...", proc.cam.Title())
	proc.streamProcess.Stop()
}

func (proc *persistCameraToDisk) Wait() {
	log.Debug("Waiting for monitoring camera on/off state change to shutdown...")
	proc.monitorCameraOnState.Wait()
	log.Info("Waiting for writing clips to disk shutdown...")
	proc.persistClips.Wait()
	log.Info("Waiting for generating clips to shutdown...")
	proc.generateClips.Wait()
	log.Info("Waiting for streaming video to shutdown...")
	proc.streamProcess.Wait()
}

func sendEvtOnCameraStateChange(b *broadcast.Broadcaster, conn camera.Connection) func(context.Context) []chan interface{} {
	return func(c context.Context) []chan interface{} {
		stopping := make(chan interface{})
		t := time.NewTicker(1 * time.Second)
		wasOff := false
	procLoop:
		for {
			time.Sleep(1 * time.Microsecond)
			select {
			case <-c.Done():
				t.Stop()
				close(stopping)
				break procLoop
			case <-t.C:
				if conn.Schedule().IsOn(schedule.Time(time.Now())) {
					if wasOff {
						b.Send(CAM_SWITCHED_ON_EVT)
					}
					wasOff = false
				} else {
					if !wasOff {
						b.Send(CAM_SWITCHED_OFF_EVT)
						wasOff = true
					}
				}
			default:
			}
		}
		return []chan interface{}{stopping}
	}
}
