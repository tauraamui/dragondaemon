package process

import (
	"context"
	"time"

	"github.com/tauraamui/dragondaemon/pkg/camera"
	"github.com/tauraamui/dragondaemon/pkg/log"
)

func DeleteOldClips(cam camera.Connection) func(cancel context.Context) []chan interface{} {
	var lastDeleteInvokedAt time.Time
	return func(cancel context.Context) []chan interface{} {
		var stopSignals []chan interface{}
		log.Info("Deleting old saved clips for camera [%s]", cam.Title())
		stopping := make(chan interface{})
		go func(cancel context.Context, cam camera.Connection, stopping chan interface{}) {
		procLoop:
			for {
				time.Sleep(1 * time.Microsecond)
				select {
				case <-cancel.Done():
					close(stopping)
					break procLoop
				default:
					lastDeleteInvokedAt = delete(cam, lastDeleteInvokedAt)
				}
			}
		}(cancel, cam, stopping)
		stopSignals = append(stopSignals, stopping)
		return stopSignals
	}
}

func delete(cam camera.Connection, lastRun time.Time) time.Time {
	if time.Now().After(lastRun.Add(5 * time.Minute)) {
		println("running delete")
		return time.Now()
	}
	return lastRun
}
