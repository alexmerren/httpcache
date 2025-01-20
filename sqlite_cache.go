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
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	createDatabaseQuery = "CREATE TABLE IF NOT EXISTS responses (request_url TEXT PRIMARY KEY, request_method TEXT, request_body TEXT, response_body TEXT, status_code INTEGER)"
	saveRequestQuery    = "INSERT INTO responses (request_url, request_method, request_body, response_body, status_code) VALUES (?, ?, ?, ?, ?)"
	readRequestQuery    = "SELECT response_body, status_code FROM responses WHERE request_url = ? AND request_method = ? AND request_body = ?"
)

// Add doc
type SqliteCache struct {
	database *sql.DB
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
		database: conn,
	}, nil
}

func (s *SqliteCache) Save(response *http.Response) error {
	return s.SaveContext(context.Background(), response)
}

func (s *SqliteCache) SaveContext(ctx context.Context, response *http.Response) error {
	requestURL := response.Request.URL.Host + response.Request.URL.RequestURI()
	requestMethod := response.Request.Method
	requestBody := []byte{}

	// If the request has no body (e.g. GET request) then Request.GetBody will
	// be nil. Check if it is nil to get request body for other types of request
	// (e.g. POST, PATCH).
	if response.Request.GetBody != nil {
		body, err := response.Request.GetBody()
		if err != nil {
			return err
		}
		requestBody, err = io.ReadAll(body)
		if err != nil {
			return err
		}
		defer body.Close()
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// Reset the response body stream to the beginning to be read again.
	response.Body = io.NopCloser(bytes.NewReader(responseBody))
	responseStatusCode := response.StatusCode

	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(saveRequestQuery, requestURL, requestMethod, requestBody, responseBody, responseStatusCode)
	if err != nil {
		return err
	}

	return err
}

func (s *SqliteCache) Read(request *http.Request) (*http.Response, error) {
	return s.ReadContext(context.Background(), request)
}

func (s *SqliteCache) ReadContext(ctx context.Context, request *http.Request) (*http.Response, error) {
	requestURL := request.URL.Host + request.URL.RequestURI()
	requestMethod := request.Method
	requestBody := []byte{}

	// If the request has no body (e.g. GET request) then Request.GetBody will
	// be nil. Check if it is nil to get request body for other types of request
	// (e.g. POST, PATCH).
	if request.GetBody != nil {
		body, err := request.GetBody()
		if err != nil {
			return nil, err
		}
		requestBody, err = io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		defer body.Close()
	}

	responseBody := ""
	responseStatusCode := 0

	if s.database == nil {
		return nil, fmt.Errorf("database is nil")
	}

	row := s.database.QueryRow(readRequestQuery, requestURL, requestMethod, requestBody)
	err := row.Scan(&responseBody, &responseStatusCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoResponse
		}
		return nil, err
	}

	return &http.Response{
		Request:    request,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
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
