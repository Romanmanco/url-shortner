package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"url-shortner/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const fn = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	//set long url address
	//create table where save info
	//after get short url address

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS url(
	    id INTEGER PRIMARY KEY,
	    alias TEXT NOT NULL UNIQUE,
	    url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);      
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToDb, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare(`
	INSERT INTO url(url, alias)
	VALUES (?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(urlToDb, alias)
	if err != nil {
		var dbErr sqlite3.Error
		// TODO: refactoring
		if errors.As(err, &dbErr) && dbErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	idURL, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return idURL, nil
}
