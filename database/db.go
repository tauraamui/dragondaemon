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

const (
	configDirType    = configdir.Global
	vendorName       = "tacusci"
	appName          = "dragondaemon"
	databaseFileName = "dd.db"
)

var (
	ErrCreateDBFile    = errors.New("unable to create database file")
	ErrDBAlreadyExists = errors.New("database file already exists")

	configDir configdir.ConfigDir
)

func init() {
	configDir = configdir.New(vendorName, appName)
}

func Setup() error {
	logging.Info("Creating database file...")

	if err := createFile(); err != nil {
		return err
	}

	// TODO(tauraamui): Create DB connection pointer here

	fmt.Println("Please enter root admin credentials...")
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
	dbFilePath, err := resolveDBPath()
	if err != nil {
		return fmt.Errorf("unable to delete database file: %w", err)
	}

	return os.Remove(dbFilePath)
}

func Connect() (*gorm.DB, error) {
	return nil, nil
}

func resolveDBPath() (string, error) {
	dbParentDir := configDir.QueryFolderContainsFile(databaseFileName)
	if dbParentDir == nil {
		return "", fmt.Errorf("unable to find %s in config location", databaseFileName)
	}
	return fmt.Sprintf("%s%c%s", dbParentDir.Path, os.PathSeparator, databaseFileName), nil
}

func createRootUser(db *gorm.DB, username, password string) error {
	return nil
}

func createFile() error {
	folder := configDir.QueryFolderContainsFile(databaseFileName)
	if folder == nil {
		folders := configDir.QueryFolders(configDirType)
		_, err := folders[0].Create(databaseFileName)
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
