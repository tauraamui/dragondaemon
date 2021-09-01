package process

import (
	"context"
	"os"
	"runtime"
	"strings"
	"time"

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
	cam                 camera.Connection
	writer              video.ClipWriter
	frames              chan video.Frame
	clips               chan video.Clip
	streamProcess       Process
	generateClips       Process
	persistClips        Process
	runtimeStatsEnabled bool
	outputRuntimeStats  Process
}

func (proc *persistCameraToDisk) Setup() {
	runtimeStatsEnv := strings.ToLower(os.Getenv("DRAGON_RUNTIME_STATS"))
	if runtimeStatsEnv == "1" || runtimeStatsEnv == "true" || runtimeStatsEnv == "yes" {
		proc.runtimeStatsEnabled = true
	}
	proc.streamProcess = NewStreamConnProcess(proc.cam, proc.frames)
	proc.generateClips = NewGenerateClipProcess(proc.frames, proc.clips, proc.cam.FPS()*proc.cam.SPC(), proc.cam.FullPersistLocation())
	proc.persistClips = NewPersistClipProcess(proc.clips, proc.writer)
	if proc.runtimeStatsEnabled {
		outputRuntimeStatsProcess := Settings{
			WaitForShutdownMsg: "",
			Process:            outputRuntimeStats(),
		}
		proc.outputRuntimeStats = New(outputRuntimeStatsProcess)
	}
}

func outputRuntimeStats() func(context.Context) []chan interface{} {
	return func(cancel context.Context) []chan interface{} {
		var stopSignals []chan interface{}
		stopping := make(chan interface{})
	procLoop:
		for {
			time.Sleep(1 * time.Second)
			select {
			case <-cancel.Done():
				close(stopping)
				break procLoop
			default:
				stats := runtime.MemStats{}
				runtime.ReadMemStats(&stats)
				renderStats(stats)
			}
		}
		stopSignals = append(stopSignals, stopping)
		return stopSignals
	}
}

const KB float64 = 1024
const MB = KB * KB
const GB = MB * MB

func renderStats(stats runtime.MemStats) {
	unit := MB
	outputFormat := "\n--------------------\nGO ROUTINES: %d\nALLOC: %f %s\nTOTAL ALLOC: %f %s\nSYS: %f %s\nMALLOCS: %d\nFREES: %d\nLIVE OBJS: %d"
	log.Info(
		outputFormat,
		runtime.NumGoroutine(),
		float64(stats.Alloc)/unit, resolveUnitLabel(unit),
		float64(stats.TotalAlloc)/unit, resolveUnitLabel(unit),
		float64(stats.Sys)/unit, resolveUnitLabel(unit),
		stats.Mallocs, stats.Frees,
		stats.Mallocs-stats.Frees,
	)
}

func resolveUnitLabel(unit float64) string {
	if unit == KB {
		return "KB"
	}
	if unit == MB {
		return "MB"
	}
	if unit == GB {
		return "GB"
	}
	return "N/A"
}

func (proc *persistCameraToDisk) Start() {
	if proc.runtimeStatsEnabled {
		go proc.outputRuntimeStats.Start()
	}
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
	if proc.runtimeStatsEnabled {
		proc.outputRuntimeStats.Stop()
	}
	log.Info("Stopping writing clips to disk from camera [%s] video stream...", proc.cam.Title())
	proc.persistClips.Stop()
	log.Info("Stopping generating clips from camera [%s] video stream...", proc.cam.Title())
	proc.generateClips.Stop()
	log.Info("Closing camera [%s] video stream...", proc.cam.Title())
	proc.streamProcess.Stop()
}

func (proc *persistCameraToDisk) Wait() {
	if proc.runtimeStatsEnabled {
		proc.outputRuntimeStats.Wait()
	}
	log.Info("Waiting for writing clips to disk shutdown...")
	proc.persistClips.Wait()
	log.Info("Waiting for generating clips to shutdown...")
	proc.generateClips.Wait()
	log.Info("Waiting for streaming video to shutdown...")
	proc.streamProcess.Wait()
}
