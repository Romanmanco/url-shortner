package postgres

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

// before run test make sure that you have db and conditionals
const dbTestURL = "user=postgres dbname=url-shortner-test password=password sslmode=disable"

const testAlias = "test_alias"
const testURL = "http://example.com"

// before test init db
func InitDb(t *testing.T) *sql.DB {
	// connect to test db
	db, err := sql.Open("postgres", dbTestURL)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	return db
}

func TestNew(t *testing.T) {
	testDb := InitDb(t)

	defer func(testDb *sql.DB) {
		err := testDb.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(testDb)

	// create obj bd with new
	storage, err := New(dbTestURL)
	assert.NoError(t, err, "Failed to create storage")

	// check that db exists
	_, err = testDb.Exec("SELECT id FROM url LIMIT 1")
	assert.NoError(t, err, "Table 'url' does not exist in the database")

	// check that db not empty
	assert.NotNil(t, storage, "Storage object is nil")
}

func TestStorage_SaveURL(t *testing.T) {
	testDb := InitDb(t)

	defer func(testDb *sql.DB) {
		err := testDb.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(testDb)

	// create db obj
	storage := &Storage{db: testDb}

	// clean table
	_, err := testDb.Exec("DELETE FROM url")
	assert.NoError(t, err, "Failed to clear the database")

	// add test data to save, checking that id is positive
	id, err := storage.SaveURL(testURL, testAlias)
	assert.NoError(t, err, "Failed to save URL")
	assert.NotNil(t, id, "Expected a positive ID")

	// checking that data is correct
	var url string
	err = testDb.QueryRow("SELECT url FROM url WHERE alias = $1", testAlias).Scan(&url)
	assert.NoError(t, err, "Failed to save URL")
	assert.Equal(t, testURL, "http://example.com")
}

func TestStorage_GetURL(t *testing.T) {
	testDb := InitDb(t)

	defer func(testDb *sql.DB) {
		err := testDb.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(testDb)

	// create db obj
	storage := &Storage{db: testDb}

	// clean table before checks
	_, err := testDb.Exec("DELETE FROM url")
	assert.NoError(t, err, "Failed to clear the database")

	// add test data before check that get url is work
	_, err = storage.SaveURL(testURL, testAlias)
	assert.NoError(t, err, "Failed to save URL")

	// get saved url
	url, err := storage.GetURL(testAlias)
	assert.NoError(t, err, "Failed to get URL")
	assert.NotEmpty(t, url, "Expected not empty URL")

	// checking that data is correct
	err = testDb.QueryRow("SELECT url FROM url WHERE alias = $1", testAlias).Scan(&url)
	assert.NoError(t, err, "Failed to get URL")
	assert.Equal(t, testURL, "http://example.com")
}

func TestStorage_DeleteURL(t *testing.T) {

}

func TestStorage_ShowAllURLs(t *testing.T) {

}

func TestStorage_UpdateURL(t *testing.T) {

}
