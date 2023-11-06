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

	// insert in database
	_, err := s.db.Exec("INSERT INTO url(url, alias) VALUES ($1, $2)", urlToDb, alias)
	if err != nil {
		// checking unique key
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, storage.ErrURLExists
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// get last id
	var idURL int64
	err = s.db.QueryRow("SELECT id FROM url WHERE url = $1 AND alias = $2", urlToDb, alias).Scan(&idURL)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return idURL, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	var url string
	err := s.db.QueryRow("SELECT url FROM url WHERE alias = $1", alias).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return url, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.postgres.DeleteURL"

	_, err := s.db.Exec("DELETE FROM url WHERE alias = $1", alias)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrURLNotFound
		}
		return fmt.Errorf("%s: failed to deletind by alias: %w", op, err)
	}

	return fmt.Errorf("%s: success deletind by alias", op)
}

func (s *Storage) ShowAllURLs() ([]storage.URLInfo, error) {
	const op = "storage.postgres.ShowAllURLs"

	rows, err := s.db.Query("SELECT alias, url FROM url")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var urls []storage.URLInfo
	for rows.Next() {
		var info storage.URLInfo
		if err := rows.Scan(&info.Alias, &info.URL); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		urls = append(urls, info)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return urls, nil
}

func (s *Storage) UpdateURL(alias, newURL string) error {
	const op = "storage.postgres.UpdateURL"

	_, err := s.db.Exec("UPDATE url SET url = $1 WHERE alias = $2", newURL, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// check that rows updated
	rowsAffected, err := s.getRowsAffected("UPDATE url SET url = $1 WHERE alias = $2", newURL, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// check that alias found
	if rowsAffected == 0 {
		return storage.ErrURLNotFound
	}

	return fmt.Errorf("%s: success update by alias", op)
}

func (s *Storage) getRowsAffected(query string, args ...interface{}) (int64, error) {
	res, err := s.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
