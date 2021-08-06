package videotest

import (
	"os"
	"path/filepath"
)

func RestoreMp4File() (string, error) {
	mp4Dir := os.TempDir()
	err := RestoreAsset(mp4Dir, "small.mp4")
	if err != nil {
		return "", err
	}

	return filepath.Join(mp4Dir, "small.mp4"), nil
}
