package videostorage

import (
	"database/sql"
	"time"

	"github.com/nakabonne/tstorage"
	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"

	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	SaveFrames(time int64, frame []videoframe.Frame) error
}

func NewStorage() (Storage, error) {
	return newSQLite3Storage()
}

type sqlite3Storage struct {
	db *sql.DB
}

func newSQLite3Storage() (*sqlite3Storage, error) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		return nil, err
	}

	s := sqlite3Storage{db}
	if err := s.init(); err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *sqlite3Storage) init() error {
	_, err := s.db.Exec(`
		CREATE TABLE autoinc(num INTEGER); INSERT INTO autoinc(num) VALUES(0);
		CREATE TABLE data(dt INTEGER, id INTEGER, data BLOB, PRIMARY KEY(dt, id)) WITHOUT ROWID;
		CREATE TRIGGER insert_trigger BEFORE INSERT ON data BEGIN UPDATE autoinc SET num=num+1; END;
	`)

	return err
}

func (s *sqlite3Storage) SaveFrames(time int64, frames []videoframe.Frame) error {
	r, err := s.db.Exec("INSERT INTO data(dt, id, data) VALUES (?, (SELECT num FROM autoinc), ?);", time, []byte{0x33})
	if err != nil {
		return err
	}

	return nil
}

type tstorageBackend struct {
	store tstorage.Storage
}

func newTStorage() (*tstorageBackend, error) {
	s, err := tstorage.NewStorage(tstorage.WithPartitionDuration(time.Microsecond))
	if err != nil {
		return nil, err
	}

	return &tstorageBackend{s}, nil
}

func (t *tstorageBackend) SaveFrame(frame videoframe.Frame) error {
	time := frame.Timestamp()
	rows := []tstorage.Row{}

	rows = append(rows, tstorage.Row{DataPoint: tstorage.DataPoint{
		Timestamp: time,
	}})

	return nil
}
