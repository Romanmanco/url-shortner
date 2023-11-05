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

	// Создаем таблицу для сохранения информации
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

	// Подготовка SQL-запроса с использованием $1 и $2 для параметров.
	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES ($1, $2)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Выполнение SQL-запроса с параметрами urlToDb и alias.
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
