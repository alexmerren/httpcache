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

// Add doc
type SqliteCache struct {
	Database *sql.DB
}

// Add doc
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

func (s *SqliteCache) SaveContext(ctx context.Context, response *http.Response) error {
	requestUrl := response.Request.URL.Host + response.Request.URL.RequestURI()
	requestMethod := response.Request.Method

	responseBody := bytes.NewBuffer(nil)
	_, err := io.Copy(responseBody, response.Body)
	if err != nil {
		return err
	}

	// Reset the response body stream to the beginning to be read again.
	response.Body = io.NopCloser(responseBody)
	responseStatusCode := response.StatusCode

	_, err = s.Database.Exec(saveRequestQuery, requestUrl, requestMethod, responseBody.Bytes(), responseStatusCode)
	if err != nil {
		return err
	}

	return err
}

func (s *SqliteCache) Read(request *http.Request) (*http.Response, error) {
	return s.ReadContext(context.Background(), request)
}

func (s *SqliteCache) ReadContext(ctx context.Context, request *http.Request) (*http.Response, error) {
	if s.Database == nil {
		return nil, fmt.Errorf("database is nil")
	}

	requestURL := request.URL.Host + request.URL.RequestURI()
	requestMethod := request.Method
	row := s.Database.QueryRow(readRequestQuery, requestURL, requestMethod)

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
