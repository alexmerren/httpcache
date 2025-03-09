package httpcache_test

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexmerren/httpcache"
	"github.com/stretchr/testify/assert"
)

const (
	testHost = "www.test.com"
	testPath = "/test-path"
	testBody = "this is a test body"
)

const (
	insertQuery = `INSERT OR REPLACE INTO responses \( request_url, request_method, response_body, status_code, expiry_time\) VALUES \(\?, \?, \?, \?, \?\)`
	selectQuery = `SELECT response_body, status_code, expiry_time FROM responses WHERE request_url = \? AND request_method = \?`
)

func Test_Save_HappyPath(t *testing.T) {
	// Given
	response := aDummyResponse()
	db, mockDatabase, closeFunc := aDatabaseMock(t)
	defer closeFunc()

	mockDatabase.
		ExpectExec(insertQuery).
		WithArgs(
			testHost+testPath,
			http.MethodGet,
			[]byte(testBody),
			http.StatusOK,
			nil,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	subject := &httpcache.SqliteCache{Database: db}

	// When
	err := subject.Save(response, nil)

	// Then
	assert.Nil(t, err)
	if err := mockDatabase.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_Save_HappyPathWithExpiryTime(t *testing.T) {
	// Given
	expiryTimestamp := time.Duration(1) * time.Minute

	response := aDummyResponse()
	db, mockDatabase, closeFunc := aDatabaseMock(t)
	defer closeFunc()

	mockDatabase.
		ExpectExec(insertQuery).
		WithArgs(
			testHost+testPath,
			http.MethodGet,
			[]byte(testBody),
			http.StatusOK,
			time.Now().Add(expiryTimestamp).Unix(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	subject := &httpcache.SqliteCache{Database: db}

	// When
	err := subject.Save(response, &expiryTimestamp)

	// Then
	assert.Nil(t, err)
	if err := mockDatabase.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_Save_ExecFails(t *testing.T) {
	// Given
	expiryTimestamp := time.Duration(1) * time.Minute

	response := aDummyResponse()
	db, mockDatabase, closeFunc := aDatabaseMock(t)
	defer closeFunc()

	mockDatabase.
		ExpectExec(insertQuery).
		WithArgs(
			testHost+testPath,
			http.MethodGet,
			[]byte(testBody),
			http.StatusOK,
			time.Now().Add(expiryTimestamp).Unix(),
		).
		WillReturnError(errors.New("dummy error"))

	subject := &httpcache.SqliteCache{Database: db}

	// When
	err := subject.Save(response, &expiryTimestamp)

	// Then
	assert.NotNil(t, err)
	if err := mockDatabase.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_Read_HappyPath(t *testing.T) {
	// Given
	request := aDummyRequest()
	db, mockDatabase, closeFunc := aDatabaseMock(t)
	defer closeFunc()

	mockRows := sqlmock.
		NewRows([]string{"response_body", "status_code", "expiry_time"}).
		AddRow([]byte(testBody), 200, nil)

	mockDatabase.
		ExpectQuery(selectQuery).
		WithArgs(testHost+testPath, http.MethodGet).
		WillReturnRows(mockRows)

	subject := &httpcache.SqliteCache{Database: db}

	// When
	response, err := subject.Read(request)

	// Then
	assert.Nil(t, err)
	assert.NotNil(t, response)

	responseBody, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	assert.Nil(t, err)
	assert.Equal(t, string(responseBody), testBody)

	if err := mockDatabase.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_Read_HappyPathWithExpiryTime(t *testing.T) {
	// Given
	expiryTimestamp := time.Duration(1) * time.Minute

	request := aDummyRequest()
	db, mockDatabase, closeFunc := aDatabaseMock(t)
	defer closeFunc()

	mockRows := sqlmock.
		NewRows([]string{"response_body", "status_code", "expiry_time"}).
		AddRow([]byte(testBody), 200, time.Now().Add(expiryTimestamp).Unix())

	mockDatabase.
		ExpectQuery(selectQuery).
		WithArgs(testHost+testPath, http.MethodGet).
		WillReturnRows(mockRows)

	subject := &httpcache.SqliteCache{Database: db}

	// When
	response, err := subject.Read(request)

	// Then
	assert.Nil(t, err)
	assert.NotNil(t, response)

	responseBody, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	assert.Nil(t, err)
	assert.Equal(t, string(responseBody), testBody)

	if err := mockDatabase.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_Read_ExpiryTimeHasPassed(t *testing.T) {
	// Given
	expiryTimestamp := -time.Duration(10) * time.Minute

	request := aDummyRequest()
	db, mockDatabase, closeFunc := aDatabaseMock(t)
	defer closeFunc()

	mockRows := sqlmock.
		NewRows([]string{"response_body", "status_code", "expiry_time"}).
		AddRow([]byte(testBody), 200, time.Now().Add(expiryTimestamp).Unix())

	mockDatabase.
		ExpectQuery(selectQuery).
		WithArgs(testHost+testPath, http.MethodGet).
		WillReturnRows(mockRows)

	subject := &httpcache.SqliteCache{Database: db}

	// When
	response, err := subject.Read(request)

	// Then
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, httpcache.ErrNoResponse)
	assert.Nil(t, response)

	if err := mockDatabase.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_Read_NoRowsFromScan(t *testing.T) {
	// Given
	request := aDummyRequest()
	db, mockDatabase, closeFunc := aDatabaseMock(t)
	defer closeFunc()

	mockDatabase.
		ExpectQuery(selectQuery).
		WithArgs(testHost+testPath, http.MethodGet).
		WillReturnError(sql.ErrNoRows)

	subject := &httpcache.SqliteCache{Database: db}

	// When
	response, err := subject.Read(request)

	// Then
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, httpcache.ErrNoResponse)
	assert.Nil(t, response)

	if err := mockDatabase.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_Read_ScanError(t *testing.T) {
	// Given
	request := aDummyRequest()
	db, mockDatabase, closeFunc := aDatabaseMock(t)
	defer closeFunc()

	mockDatabase.
		ExpectQuery(selectQuery).
		WithArgs(testHost+testPath, http.MethodGet).
		WillReturnError(sql.ErrConnDone)

	subject := &httpcache.SqliteCache{Database: db}

	// When
	response, err := subject.Read(request)

	// Then
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, sql.ErrConnDone)
	assert.Nil(t, response)

	if err := mockDatabase.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func aDatabaseMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func() error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("%v", err)
	}

	return db, mock, db.Close
}

func aDummyResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(testBody)),
		Request: &http.Request{
			URL: &url.URL{
				Scheme:      "https",
				Host:        testHost,
				Opaque:      "",
				User:        nil,
				Path:        testPath,
				RawPath:     "",
				OmitHost:    false,
				ForceQuery:  false,
				RawQuery:    "",
				Fragment:    "",
				RawFragment: "",
			},
			Method: http.MethodGet,
		},
	}
}

func aDummyRequest() *http.Request {
	return &http.Request{
		URL: &url.URL{
			Scheme:      "https",
			Host:        testHost,
			Opaque:      "",
			User:        nil,
			Path:        testPath,
			RawPath:     "",
			OmitHost:    false,
			ForceQuery:  false,
			RawQuery:    "",
			Fragment:    "",
			RawFragment: "",
		},
		Method: http.MethodGet,
	}
}
