package data

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/shibukawa/configdir"
	"github.com/spf13/afero"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/database/models"
	"github.com/tauraamui/dragondaemon/pkg/database/repos"
	"golang.org/x/term"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	configDirType    = configdir.System
	vendorName       = "tacusci"
	appName          = "dragondaemon"
	databaseFileName = "dd.db"
)

var (
	ErrCreateDBFile    = errors.New("unable to create database file")
	ErrDBAlreadyExists = errors.New("database file already exists")
)

var uc = os.UserCacheDir
var fs = afero.NewOsFs()
var promptReader io.Reader = os.Stdin
var passwordPromptReader passwordReader = stdinPasswordReader{}

type passwordReader interface {
	ReadPassword() ([]byte, error)
}

type stdinPasswordReader struct{}

func (s stdinPasswordReader) ReadPassword() ([]byte, error) {
	return term.ReadPassword(syscall.Stdin)
}

func Setup() error {
	logging.Info("Creating database file...")

	if err := createFile(uc, fs); err != nil {
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
	dbFilePath, err := resolveDBPath(uc)
	if err != nil {
		return fmt.Errorf("unable to delete database file: %w", err)
	}

	return os.Remove(dbFilePath)
}

func Connect() (*gorm.DB, error) {
	dbPath, err := resolveDBPath(uc)
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
	userRepo := repos.UserRepository{DB: db}
	return userRepo.Create(&models.User{
		Name:     username,
		AuthHash: password,
	})
}

func resolveDBPath(uc func() (string, error)) (string, error) {
	databasePath := os.Getenv("DRAGON_DAEMON_DB")
	if len(databasePath) > 0 {
		return databasePath, nil
	}

	databaseParentDir, err := uc()
	if err != nil {
		return "", fmt.Errorf("unable to resolve %s database file location: %w", databaseFileName, err)
	}

	return filepath.Join(
		databaseParentDir,
		vendorName,
		appName,
		databaseFileName), nil
}

func createFile(uc func() (string, error), fs afero.Fs) error {
	path, err := resolveDBPath(uc)
	if err != nil {
		return err
	}

	if _, err := fs.Stat(path); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(strings.Replace(path, databaseFileName, "", -1), 0700)
		_, err := fs.Create(path)
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
	stdinReader := bufio.NewReader(promptReader)
	value, err := stdinReader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(value), nil
}

func promptForValueEchoOff(promptText string) (string, error) {
	fmt.Printf("%s: ", promptText)
	valueBytes, err := passwordPromptReader.ReadPassword()
	if err != nil {
		return "", err
	}
	fmt.Println("")
	return strings.TrimSpace(string(valueBytes)), nil
}
