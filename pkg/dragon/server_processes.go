package dragon

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/dragon/process"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"gocv.io/x/gocv"
)

func (s *Server) SetupProcesses() {
	s.initDebugProcs()
	for _, cam := range s.cameras {
		proc := process.NewCoreProcess(cam, s.videoBackend.NewWriter())
		proc.Setup()
		s.coreProcesses[cam.UUID()] = proc
	}
}

func (s *Server) initDebugProcs() {
	runtimeStatsEnv := strings.ToLower(os.Getenv("DRAGON_RUNTIME_STATS"))
	if runtimeStatsEnv == "1" || runtimeStatsEnv == "true" || runtimeStatsEnv == "yes" {
		s.runtimeStatsEnabled = true
		outputRuntimeStatsProcess := process.Settings{
			WaitForShutdownMsg: "",
			Process:            outputRuntimeStats(),
		}
		s.renderRuntimeStatsProc = process.New(outputRuntimeStatsProcess)
	}

	openCVMatStatsEnv := strings.ToLower(os.Getenv("DRAGON_OPENCV_MAT_STATS"))
	if openCVMatStatsEnv == "1" || openCVMatStatsEnv == "true" || openCVMatStatsEnv == "yes" {
		s.openCVMatStatsEnabled = true
		outputOpenCVStatsProcess := process.Settings{
			WaitForShutdownMsg: "",
			Process:            outputOpenCVMatStats(),
		}
		s.renderOpenCVMatStatsProc = process.New(outputOpenCVStatsProcess)
	}
}

func outputOpenCVMatStats() func(context.Context, chan interface{}) []chan interface{} {
	return func(cancel context.Context, s chan interface{}) []chan interface{} {
		stopping := make(chan interface{})
		started := false

	procLoop:
		for {
			time.Sleep(5 * time.Second)
			if !started {
				close(s)
				started = true
			}
			select {
			case <-cancel.Done():
				close(stopping)
				break procLoop
			default:
				var b bytes.Buffer
				gocv.MatProfile.Count()
				gocv.MatProfile.WriteTo(&b, 1)
				fmt.Print(b.String())
			}
		}

		return []chan interface{}{stopping}
	}
}

func outputRuntimeStats() func(context.Context, chan interface{}) []chan interface{} {
	return func(cancel context.Context, s chan interface{}) []chan interface{} {
		stopping := make(chan interface{})
		started := false
	procLoop:
		for {
			time.Sleep(1 * time.Second)
			if !started {
				close(s)
				started = true
			}
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

func (s *Server) RunProcesses() {
	if s.runtimeStatsEnabled && s.renderRuntimeStatsProc != nil {
		s.renderRuntimeStatsProc.Start()
	}

	if s.openCVMatStatsEnabled && s.renderOpenCVMatStatsProc != nil {
		s.renderOpenCVMatStatsProc.Start()
	}

	for _, proc := range s.coreProcesses {
		proc.Start()
	}
}

func (s *Server) shutdownProcesses() {
	if s.runtimeStatsEnabled && s.renderRuntimeStatsProc != nil {
		s.renderRuntimeStatsProc.Stop()
		s.renderRuntimeStatsProc.Wait()
	}

	if s.openCVMatStatsEnabled && s.renderOpenCVMatStatsProc != nil {
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
