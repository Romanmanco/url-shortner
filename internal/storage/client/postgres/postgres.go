package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"url-shortner/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(dbURL string) (*Storage, error) {
	const fn = "storage.postgres.New"

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	// create table for save info
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS url (
	    id SERIAL PRIMARY KEY,
	    alias TEXT NOT NULL UNIQUE,
	    url TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToDb, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"

	// prepare sql for save with $1 и $2 params
	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES ($1, $2)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// make request to db
	_, err = stmt.Exec(urlToDb, alias)
	if err != nil {
		// checking unique key
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// get last id
	var idURL int64
	err = s.db.QueryRow("INSERT INTO url(url, alias) VALUES ($1, $2) RETURNING id", urlToDb, alias).Scan(&idURL)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return idURL, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	// prepare sql for searching url by alias
	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// get url by alias
	var fullUrl string
	err = stmt.QueryRow(alias).Scan(&fullUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: failed to get url by alias: %w", op, err)
	}

	return fullUrl, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.postgres.DeleteURL"

	// prepare sql for deleting url by alias
	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = stmt.QueryRow(alias).Scan()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrURLNotFound
		}
		return fmt.Errorf("%s: failed to deletind by alias: %w", op, err)
	}

	return fmt.Errorf("%s: success deletind by alias", op)
}
