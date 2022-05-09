package videostorage

import (
	"bytes"
	"database/sql"
	"fmt"

	"github.com/tauraamui/dragondaemon/pkg/video/videoframe"

	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	SaveFrames(time int64, frame []videoframe.Frame) error
	Close() error
}

func NewStorage(path string) (Storage, error) {
	return newSQLite3Storage(path)
}

const SQLITE_INMEM_FILE_PATH = "file::memory:?cache=shared"

type sqlite3Storage struct {
	db *sql.DB
}

func newSQLite3Storage(path string) (*sqlite3Storage, error) {
	db, err := sql.Open("sqlite3", path)
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
	blob, err := convertFramesToBlob(frames)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("INSERT INTO data(dt, id, data) VALUES (?, (SELECT num FROM autoinc), ?);", time, blob)
	if err != nil {
		return err
	}

	return nil
}

func (s *sqlite3Storage) Close() error {
	return s.db.Close()
}

func convertFramesToBlob(frames []videoframe.Frame) ([]byte, error) {
	buff := bytes.Buffer{}
	framesCount := len(frames)
	for i := 0; i < framesCount; i++ {
		f := frames[i]
		fb := f.ToBytes()
		wc, err := buff.Write(fb)
		if err != nil {
			return nil, fmt.Errorf("something went critically wrong, have run out of memory??: %w", err)
		}

		if fc := len(fb); fc != wc {
			return nil, fmt.Errorf("writing all of the bytes from frames failed, wrote: %d out of %d", wc, fc)
		}

		if i+1 < framesCount {
			buff.Write([]byte{0x34, 0xE7}) // the delimiter
		}
	}
	return buff.Bytes(), nil
}
