package data

import (
	"errors"
	"fmt"
	"os"

	"github.com/shibukawa/configdir"
	"github.com/tacusci/logging/v2"
)

var (
	ErrCreateDB = errors.New("unable to create dragondaemon database")
)

func Create() error {
	configDirs := configdir.New("tacusci", "dragondaemon")
	folder := configDirs.QueryFolderContainsFile("dd.db")
	if folder == nil {
		folders := configDirs.QueryFolders(configdir.Global)
		_, err := folders[0].Create("dd.db")
		if err != nil {
			return fmt.Errorf("create db file: %w", err)
		}
		return nil
	}

	return nil
}

func CreateRootUser(username, password string) error {
	logging.Info("Creating root user of name %s with password %s", username, password)
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
	return errors.New("Unable to find app resource dir to remove")
}
