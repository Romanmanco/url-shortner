package storage

import "errors"

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url exists")
)

type URLInfo struct {
	Alias string `json:"alias"`
	URL   string `json:"url"`
}
