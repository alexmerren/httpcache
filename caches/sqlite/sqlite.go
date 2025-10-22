package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alexmerren/httpcache"
	_ "github.com/mattn/go-sqlite3"
)

const (
	createDatabaseQuery = `
	CREATE TABLE IF NOT EXISTS responses (
		request_url TEXT PRIMARY KEY, 
		request_method TEXT NOT NULL, 
		response_body BLOB NOT NULL, 
		status_code INTEGER NOT NULL, 
		created_at INTEGER NOT NULL)`

	saveRequestQuery = `
	INSERT OR REPLACE INTO responses (
		request_url, 
		request_method, 
		response_body, 
		status_code, 
		created_at) 
	VALUES (?, ?, ?, ?, ?)`

	readRequestQuery = `
	SELECT response_body, status_code, created_at 
	FROM responses 
	WHERE request_url = ? AND request_method = ?`
)

// ErrNoDatabase denotes when the database does not exist or has not been
// constructed correctly.
var ErrNoDatabase = errors.New("no database connection")

// SqliteDriver is a default implementation of [Cache] which creates a local
// SQLite driver to persist and to query HTTP responses.
type SqliteDriver struct {
	Database *sql.DB
}

// NewSqliteDriver creates a new SQLite database with a certain name. This name
// is the filename of the database. If the file does not exist, then we create
// it. If the file is in a non-existent directory, we create the directory.
func NewSqliteDriver(databaseName string) (*SqliteDriver, error) {
	fileExists, err := doesFileExist(databaseName)
	if err != nil {
		return nil, err
	}

	if !fileExists {
		_, err = createFile(databaseName)
		if err != nil {
			return nil, err
		}
	}

	conn, err := sql.Open("sqlite3", databaseName)
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec(createDatabaseQuery)
	if err != nil {
		return nil, err
	}

	return &SqliteDriver{
		Database: conn,
	}, nil
}

func (s *SqliteDriver) Save(response *http.Response) error {
	return s.SaveContext(context.Background(), response)
}

func (s *SqliteDriver) Read(request *http.Request) (*http.Response, error) {
	return s.ReadContext(context.Background(), request)
}

func (s *SqliteDriver) SaveContext(ctx context.Context, response *httpcache.ReadResult) error {
	responseBody := bytes.NewBuffer(nil)
	_, err := io.Copy(responseBody, response.Body)
	if err != nil {
		return err
	}

	// Reset the response body stream to the beginning to be read again.
	response.Body = io.NopCloser(responseBody)
	now := time.Now().UnixMilli()

	requestUrl := generateUrl(response.Request)
	_, err = s.Database.Exec(
		saveRequestQuery,
		requestUrl,
		response.Request.Method,
		responseBody.Bytes(),
		response.StatusCode,
		now,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *SqliteDriver) ReadContext(ctx context.Context, request *http.Request) (*http.Response, error) {
	if s.Database == nil {
		return nil, ErrNoDatabase
	}

	requestUrl := generateUrl(request)
	row := s.Database.QueryRow(readRequestQuery, requestUrl, request.Method)

	var responseBody []byte
	var responseStatusCode int
	var createdAt *int64

	err := row.Scan(&responseBody, &responseStatusCode, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httpcache.ErrNoResult
		}
		return nil, err
	}

	return &http.Response{
		Request:    request,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
		StatusCode: responseStatusCode,
	}, nil
}

func generateUrl(request *http.Request) string {
	return request.URL.Host + request.URL.RequestURI()
}

func doesFileExist(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	return true, nil
}

func createFile(filename string) (*os.File, error) {
	directory, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return os.Create(filename)
}
