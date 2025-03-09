package httpcache

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
	createDatabaseQuery = "CREATE TABLE IF NOT EXISTS responses (request_url TEXT PRIMARY KEY, request_method TEXT, response_body BLOB, status_code INTEGER)"
	saveRequestQuery    = "INSERT INTO responses (request_url, request_method, response_body, status_code) VALUES (?, ?, ?, ?)"
	readRequestQuery    = "SELECT response_body, status_code FROM responses WHERE request_url = ? AND request_method = ?"
)

// SqliteCache is a default implementation of [Cache] which creates a local
// SQLite cache to persist and to query HTTP responses.
type SqliteCache struct {
	Database *sql.DB
}

// NewSqliteCache creates a new SQLite database with a certain name. This name
// is the filename of the database. If the file does not exist, then we create
// it. If the file is in a non-existent directory, we create the directory.
func NewSqliteCache(databaseName string) (*SqliteCache, error) {
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

	return &SqliteCache{
		Database: conn,
	}, nil
}

func (s *SqliteCache) Save(response *http.Response) error {
	return s.SaveContext(context.Background(), response)
}

func (s *SqliteCache) Read(request *http.Request) (*http.Response, error) {
	return s.ReadContext(context.Background(), request)
}

func (s *SqliteCache) SaveContext(ctx context.Context, response *http.Response) error {
	responseBody := bytes.NewBuffer(nil)
	_, err := io.Copy(responseBody, response.Body)
	if err != nil {
		return err
	}

	// Reset the response body stream to the beginning to be read again.
	response.Body = io.NopCloser(responseBody)

	requestUrl := generateUrl(response.Request)
	_, err = s.Database.Exec(saveRequestQuery, requestUrl, response.Request.Method, responseBody.Bytes(), response.StatusCode)
	if err != nil {
		return err
	}

	return nil
}

func (s *SqliteCache) ReadContext(ctx context.Context, request *http.Request) (*http.Response, error) {
	if s.Database == nil {
		return nil, fmt.Errorf("database is nil")
	}

	requestUrl := generateUrl(request)
	row := s.Database.QueryRow(readRequestQuery, requestUrl, request.Method)

	var responseBody []byte
	var responseStatusCode int
	err := row.Scan(&responseBody, &responseStatusCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoResponse
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
