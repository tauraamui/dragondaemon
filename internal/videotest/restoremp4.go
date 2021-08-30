package videotest

import (
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

func RestoreMp4File() (string, error) {
	mp4Dir := os.TempDir()
	err := RestoreAsset(mp4Dir, "small.mp4")
	if err != nil {
		return "", err
	}

	return filepath.Join(mp4Dir, "small.mp4"), nil
}

func MakeRootPath(fs afero.Fs) (string, error) {
	path := filepath.Join(os.TempDir(), "testroot")
	err := fs.MkdirAll(path, os.ModePerm|os.ModeDir)
	return path, err
}
