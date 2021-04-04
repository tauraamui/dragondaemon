package data

import (
	"errors"
	"fmt"
	"os"

	"github.com/shibukawa/configdir"
)

var (
	ErrCreateDBFile    = errors.New("unable to create database file")
	ErrDBAlreadyExists = errors.New("database file already exists")
)

func Create() error {
	configDirs := configdir.New("tacusci", "dragondaemon")
	folder := configDirs.QueryFolderContainsFile("dd.db")
	if folder == nil {
		folders := configDirs.QueryFolders(configdir.Global)
		_, err := folders[0].Create("dd.db")
		if err != nil {
			return fmt.Errorf("%v: %w", ErrCreateDBFile, err)
		}
		return nil
	}

	return ErrDBAlreadyExists
}

func CreateRootUser(username, password string) error {
	return nil
}

func Destroy() error {
	configDirs := configdir.New("tacusci", "dragondaemon")
	folder := configDirs.QueryFolderContainsFile("dd.db")
	if folder != nil {
		err := os.RemoveAll(folder.Path)
		if err != nil {
			return fmt.Errorf("unable to remove app resource dir: %w", err)
		}
		return nil
	}
	return errors.New("unable to find app resource dir to remove")
}
