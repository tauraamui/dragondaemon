package data

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/shibukawa/configdir"
	"github.com/tacusci/logging/v2"
	"golang.org/x/term"
	"gorm.io/gorm"
)

var (
	ErrCreateDBFile    = errors.New("unable to create database file")
	ErrDBAlreadyExists = errors.New("database file already exists")
)

func Setup() error {
	logging.Info("Creating database file...")

	if err := createFile(); err != nil {
		return err
	}

	logging.Info("Created database file... Please enter root admin credentials...")
	rootUsername, err := promptForValue("username", false)
	if err != nil {
		return fmt.Errorf("unable to prompt for root username: %w", err)
	}

	rootPassword, err := promptForValue("password", true)
	if err != nil {
		return fmt.Errorf("unable to prompt for root password : %w", err)
	}

	if err := createRootUser(nil, rootUsername, rootPassword); err != nil {
		return fmt.Errorf("unable to create root user entry: %w", err)
	}

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

func createRootUser(db *gorm.DB, username, password string) error {
	return nil
}

func createFile() error {
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

func promptForValue(valueName string, hidden bool) (string, error) {
	fmt.Printf("Enter %s: ", valueName)
	if hidden {
		valueBytes, err := term.ReadPassword(0)
		if err != nil {
			return "", err
		}
		fmt.Println("")
		return string(valueBytes), nil
	}
	stdinReader := bufio.NewReader(os.Stdin)
	value, err := stdinReader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return value, nil
}
