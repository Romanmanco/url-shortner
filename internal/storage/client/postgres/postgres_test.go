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
func initDb(t *testing.T) *sql.DB {
	// connect to test db
	db, err := sql.Open("postgres", dbTestURL)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	return db
}

// clean table
func cleanDb(t *testing.T, testDb *sql.DB) {
	_, err := testDb.Exec("DELETE FROM url")
	assert.NoError(t, err, "Failed to clear the database")
}

// save test data before checks
func saveDataForChecks(t *testing.T, storage *Storage, testURL, testAlias string) {
	url, err := storage.SaveURL(testURL, testAlias)
	assert.NoError(t, err, "Failed to save URL")
	assert.NotEmpty(t, url, "Expected not empty URL")
}

func TestNew(t *testing.T) {
	testDb := initDb(t)

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
	testDb := initDb(t)

	defer func(testDb *sql.DB) {
		err := testDb.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(testDb)

	// create db obj
	storage := &Storage{db: testDb}

	// clean table before ad new date
	cleanDb(t, testDb)

	// add test data to save, checking that id is positive
	saveDataForChecks(t, storage, testURL, testAlias)

	// checking that data is correct
	var url string
	err := testDb.QueryRow("SELECT url FROM url WHERE alias = $1", testAlias).Scan(&url)
	assert.NoError(t, err, "Failed to save URL")
	assert.Equal(t, testURL, "http://example.com")
}

func TestStorage_GetURL(t *testing.T) {
	testDb := initDb(t)

	defer func(testDb *sql.DB) {
		err := testDb.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(testDb)

	// create db obj
	storage := &Storage{db: testDb}

	// clean table before ad new date
	cleanDb(t, testDb)

	// add test data before check that get url is work
	saveDataForChecks(t, storage, testURL, testAlias)

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
	testDb := initDb(t)

	defer func(testDb *sql.DB) {
		err := testDb.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(testDb)

	// create db obj
	storage := &Storage{db: testDb}

	// clean table before ad new date
	cleanDb(t, testDb)

	// add test data before check that get url is work
	saveDataForChecks(t, storage, testURL, testAlias)

	// delete saved url
	err := storage.DeleteURL(testAlias)
	// get success msg
	assert.Errorf(t, err, "success deletind by alias")

	// trying get url after delete
	res, err := storage.GetURL(testAlias)
	assert.Error(t, err, "Failed to get URL")
	assert.Empty(t, res, "Expected empty URL")
}

func TestStorage_ShowAllURLs(t *testing.T) {
	testDb := initDb(t)

	defer func(testDb *sql.DB) {
		err := testDb.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(testDb)

	// create db obj
	storage := &Storage{db: testDb}

	// clean table before ad new date
	cleanDb(t, testDb)

	// add first test data before check that get url is work
	saveDataForChecks(t, storage, testURL, testAlias)

	// add second test data before check that get url is work
	saveDataForChecks(t, storage, testURL+"/second", testAlias+"2")

	// add check that show urls is work
	urls, err := storage.ShowAllURLs()
	assert.NoError(t, err, "Failed to show URL")
	assert.NotEmpty(t, urls, "Expected not empty URLs")
	assert.Equal(t, len(urls), 2, "Expected two URLs")

}

func TestStorage_UpdateURL(t *testing.T) {
	testDb := initDb(t)

	defer func(testDb *sql.DB) {
		err := testDb.Close()
		if err != nil {
			t.Fatalf("Failed to close test database: %v", err)
		}
	}(testDb)

	// create db obj
	storage := &Storage{db: testDb}

	// clean table before ad new date
	cleanDb(t, testDb)

	// add test data before check that get url is work
	saveDataForChecks(t, storage, testURL, testAlias)

	// add check that update url is work
	err := storage.UpdateURL(testAlias, testURL+"/new")
	// get success msg
	assert.Errorf(t, err, "success update by alias")

	// trying get url after update
	res, err := storage.GetURL(testAlias)
	assert.Contains(t, res, "http://example.com/new")
}
