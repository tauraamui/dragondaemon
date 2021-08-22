package process

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/afero"
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
		err := removeOldClipDirsByDate(cam.FullPersistLocation(), cam.MaxClipAgeDays())
		if err != nil {
			log.Error("error occurred whilst removing old clip dirs: %w", err)
		}
		return time.Now()
	}
	return lastRun
}

const dateLayout = "2006-01-02"

func strToDate(date string) (time.Time, error) {
	return time.Parse(dateLayout, date)
}

func removeOldClipDirsByDate(path string, maxClipAgeDays int) error {
	if err := verifyDirPath(path); err != nil {
		return err
	}

	dir, err := fs.Open(path)
	if err != nil {
		return fmt.Errorf("unable to open dir %s: %w", path, err)
	}

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		date, err := strToDate(name)
		log.Error("unable to resolve date from dir name %s: %w", name, err)
		if date.Before(time.Now().AddDate(0, 0, -1*maxClipAgeDays)) {
			if err := deleteDirAndContent(name); err != nil {
				log.Error("unable to remove dir %s: %w", name, err)
			}
		}
	}
	return nil
}

func deleteDirAndContent(path string) error {
	if err := verifyDirPath(path); err != nil {
		return err
	}
	return fs.RemoveAll(path)
}

func verifyDirPath(path string) error {

	pathIsDir, err := afero.DirExists(fs, path)
	if err != nil {
		return fmt.Errorf("unable to stat given path %s: %w", path, err)
	}

	if !pathIsDir {
		return fmt.Errorf("given path is not a directory: %s", path)
	}

	return nil
}
