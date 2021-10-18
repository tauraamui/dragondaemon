package dragon

import (
	"context"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/broadcast"
	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/config/schedule"
	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/log"
)

func (s *server) SetupProcesses() {
	runtimeStatsEnv := strings.ToLower(os.Getenv("DRAGON_RUNTIME_STATS"))
	if runtimeStatsEnv == "1" || runtimeStatsEnv == "true" || runtimeStatsEnv == "yes" {
		s.runtimeStatsEnabled = true
		outputRuntimeStatsProcess := process.Settings{
			WaitForShutdownMsg: "",
			Process:            outputRuntimeStats(),
		}
		s.renderRuntimeStatsProc = process.New(outputRuntimeStatsProcess)
	}
	for _, cam := range s.cameras {
		proc := process.NewCoreProcess(cam, s.videoBackend.NewWriter())
		proc.Setup()
		s.coreProcesses[cam.UUID()] = proc
	}
}

func monitorCameraOnState(conn camera.Connection, b *broadcast.Broadcaster) func(context.Context) []chan interface{} {
	return func(c context.Context) []chan interface{} {
		stopping := make(chan interface{})
		t := time.NewTicker(1 * time.Minute)
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
					wasOff = false
				} else {
					if !wasOff {
						b.Send(process.CAM_SWITCHED_OFF_EVT)
						wasOff = true
					}
				}
			}
		}
		return []chan interface{}{}
	}
}

func outputRuntimeStats() func(context.Context) []chan interface{} {
	return func(cancel context.Context) []chan interface{} {
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
		return []chan interface{}{stopping}
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

func (s *server) RunProcesses() {
	if s.runtimeStatsEnabled && s.renderRuntimeStatsProc != nil {
		go s.renderRuntimeStatsProc.Start()
	}
	for _, proc := range s.coreProcesses {
		proc.Start()
	}
}

func (s *server) shutdownProcesses() {
	if s.runtimeStatsEnabled && s.renderRuntimeStatsProc != nil {
		s.renderRuntimeStatsProc.Stop()
		s.renderRuntimeStatsProc.Wait()
	}
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
