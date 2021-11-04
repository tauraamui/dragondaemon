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

	"github.com/spf13/afero"
	"github.com/tacusci/logging/v2"
	"github.com/tauraamui/dragondaemon/pkg/database/dbconn"
	"github.com/tauraamui/dragondaemon/pkg/database/models"
	"github.com/tauraamui/dragondaemon/pkg/database/repos"
	"github.com/tauraamui/dragondaemon/pkg/log"
	"github.com/tauraamui/xerror"
	"golang.org/x/term"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	vendorName       = "tacusci"
	appName          = "dragondaemon"
	databaseFileName = "dd.db"
)

var (
	ErrCreateDBFile    = xerror.New("unable to create database file")
	ErrDBAlreadyExists = xerror.New("database file already exists")
)

var uc = os.UserCacheDir
var fs = afero.NewOsFs()
var plainPromptReader plainReader = stdinPlainReader{readFrom: os.Stdin}
var passwordPromptReader passwordReader = stdinPasswordReader{}

type plainReader interface {
	ReadPlain(promptText string) (string, error)
}

type passwordReader interface {
	ReadPassword(promptText string) ([]byte, error)
}

type stdinPlainReader struct {
	readFrom io.Reader
}

func (s stdinPlainReader) ReadPlain(promptText string) (string, error) {
	if len(promptText) > 0 {
		fmt.Printf("%s: ", promptText)
	}
	stdinReader := bufio.NewReader(s.readFrom)
	value, err := stdinReader.ReadString('\n')
	return strings.TrimSpace(value), err
}

type stdinPasswordReader struct{}

func (s stdinPasswordReader) ReadPassword(promptText string) ([]byte, error) {
	if len(promptText) > 0 {
		fmt.Printf("%s: ", promptText)
	}
	return term.ReadPassword(syscall.Stdin)
}

func Setup() error {
	log.Info("Creating database file...") //nolint

	if err := createFile(); err != nil {
		return err
	}

	db, err := Connect()
	if err != nil {
		return err
	}

	if logging.CurrentLoggingLevel != logging.SilentLevel {
		fmt.Println("Please enter root admin credentials...")
	}
	rootUsername, err := askForUsername()
	if err != nil {
		return xerror.Errorf("failed to prompt for root username: %w", err)
	}

	rootPassword, err := askForPassword(0)
	if err != nil {
		return xerror.Errorf("failed to prompt for root password: %w", err)
	}
	if err := createRootUser(db, rootUsername, rootPassword); err != nil {
		return xerror.Errorf("unable to create root user entry: %w", err)
	}

	log.Info("Created root admin user") //nolint

	return nil
}

func Destroy() error {
	dbFilePath, err := resolveDBPath(uc)
	if err != nil {
		return xerror.Errorf("unable to delete database file: %w", err)
	}

	return fs.Remove(dbFilePath)
}

func Connect() (dbconn.GormWrapper, error) {
	dbPath, err := resolveDBPath(uc)
	if err != nil {
		return nil, err
	}

	log.Debug("Connecting to DB: %s", dbPath) //nolint
	db, err := openDBConnection(dbPath)
	if err != nil {
		return nil, xerror.Errorf("unable to open db connection: %w", err)
	}

	err = models.AutoMigrate(db)
	if err != nil {
		return nil, xerror.Errorf("unable to run automigrations: %w", err)
	}

	return db, nil
}

var openDBConnection = func(path string) (dbconn.GormWrapper, error) {
	logger := logger.New(nil, logger.Config{LogLevel: logger.Silent})
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: logger})
	if err != nil {
		return nil, err
	}
	return dbconn.Wrap(db), nil
}

func createRootUser(db dbconn.GormWrapper, username, password string) error {
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
		return "", xerror.Errorf("unable to resolve %s database file location: %w", databaseFileName, err)
	}

	return filepath.Join(
		databaseParentDir,
		vendorName,
		appName,
		databaseFileName), nil
}

func createFile() error {
	path, err := resolveDBPath(uc)
	if err != nil {
		return err
	}

	if _, err := fs.Stat(path); errors.Is(err, os.ErrNotExist) {
		fs.MkdirAll(strings.Replace(path, databaseFileName, "", -1), os.ModeDir|os.ModePerm) //nolint

		_, err := fs.Create(path)
		if err != nil {
			return xerror.Errorf("%v: %w", ErrCreateDBFile, err)
		}
		return nil
	}

	return xerror.Errorf("%w: %s", ErrDBAlreadyExists, path)
}

func askForUsername() (string, error) {
	return promptForValue("Root admin username")
}

func askForPassword(attempts int) (string, error) {
	password, err := promptForValueEchoOff("Root user password")
	if err != nil {
		return "", xerror.Errorf("unable to prompt for root password : %w", err)
	}

	repeatedPassword, err := promptForValueEchoOff("Repeat root user password")
	if err != nil {
		return "", xerror.Errorf("unable to prompt for root password : %w", err)
	}

	if strings.Compare(password, repeatedPassword) != 0 {
		if logging.CurrentLoggingLevel != logging.SilentLevel {
			fmt.Println("Entered passwords do not match... Try again...")
		}
		attempts++
		if attempts >= 3 {
			return "", xerror.New("tried entering new password at least 3 times")
		}
		return askForPassword(attempts)
	}

	return password, nil
}

func promptForValue(promptText string) (string, error) {
	value, err := plainPromptReader.ReadPlain(promptText)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func promptForValueEchoOff(promptText string) (string, error) {
	valueBytes, err := passwordPromptReader.ReadPassword(promptText)
	if err != nil {
		return "", err
	}
	if logging.CurrentLoggingLevel != logging.SilentLevel {
		fmt.Println("")
	}
	return strings.TrimSpace(string(valueBytes)), nil
}
