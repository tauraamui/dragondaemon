package data

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/shibukawa/configdir"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/database/models"
	"golang.org/x/term"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	db, err := Connect()
	if err != nil {
		return err
	}

	fmt.Println("Please enter root admin credentials...")
	rootUsername, err := askForUsername()
	if err != nil {
		return fmt.Errorf("failed to prompt for root username: %w", err)
	}

	rootPassword, err := askForPassword(0)
	if err != nil {
		return fmt.Errorf("failed to prompt for root password: %w", err)
	}
	if err := createRootUser(db, rootUsername, rootPassword); err != nil {
		return fmt.Errorf("unable to create root user entry: %w", err)
	}

	logging.Info("Created root admin user")

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
	dbPath, err := resolveDBPath()
	logging.Debug("Connecting to DB: %s", dbPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open db connection: %w", err)
	}

	logger := logger.New(nil, logger.Config{LogLevel: logger.Silent})
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: logger})
	if err != nil {
		return nil, err
	}

	err = models.AutoMigrate(db)
	if err != nil {
		return nil, fmt.Errorf("unable to run automigrations: %w", err)
	}

	return db, nil
}

func createRootUser(db *gorm.DB, username, password string) error {
	rootUser := models.User{
		Name:     username,
		AuthHash: password,
	}

	if err := db.Create(&rootUser).Error; err != nil {
		return err
	}
	return nil
}

func resolveDBPath() (string, error) {
	dbParentDir := configDir.QueryFolderContainsFile(databaseFileName)
	if dbParentDir == nil {
		return "", fmt.Errorf("unable to find %s in config location", databaseFileName)
	}
	return fmt.Sprintf("%s%c%s", dbParentDir.Path, os.PathSeparator, databaseFileName), nil
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

func askForUsername() (string, error) {
	return promptForValue("Root admin username")
}

func askForPassword(attempts int) (string, error) {
	password, err := promptForValueEchoOff("Root user password")
	if err != nil {
		return "", fmt.Errorf("unable to prompt for root password : %w", err)
	}

	repeatedPassword, err := promptForValueEchoOff("Repeat root user password")
	if err != nil {
		return "", fmt.Errorf("unable to prompt for root password : %w", err)
	}

	if strings.Compare(password, repeatedPassword) != 0 {
		fmt.Println("Entered passwords do not match... Try again...")
		attempts++
		if attempts >= 3 {
			return "", errors.New("tried entering new password at least 3 times")
		}
		return askForPassword(attempts)
	}

	return password, nil
}

func promptForValue(promptText string) (string, error) {
	fmt.Printf("%s: ", promptText)
	stdinReader := bufio.NewReader(os.Stdin)
	value, err := stdinReader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(value), nil
}

func promptForValueEchoOff(promptText string) (string, error) {
	fmt.Printf("%s: ", promptText)
	valueBytes, err := term.ReadPassword(0)
	if err != nil {
		return "", err
	}
	fmt.Println("")
	return strings.TrimSpace(string(valueBytes)), nil
}
