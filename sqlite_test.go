package httpcache

import (
	"database/sql"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

const (
	testHost = "www.test.com"
	testPath = "/test-path"
	testBody = "this is a test body"
)

const (
	insertQuery = `INSERT INTO responses \(request_url, request_method, request_body, response_body, status_code\) VALUES \(\?, \?, \?, \?, \?\)`
	selectQuery = `SELECT response_body, status_code FROM responses WHERE request_url = \? AND request_method = \? AND request_body = \?`
)

func Test_Save_HappyPath(t *testing.T) {
	// Given
	response := aDummyResponse()
	db, mock, closeFunc := aDatabaseMock(t)
	defer closeFunc()

	subject := &sqliteResponseStore{database: db}

	mock.ExpectBegin()
	mock.ExpectExec(insertQuery).WithArgs(
		testHost+testPath,
		http.MethodGet,
		[]byte{},
		[]byte(testBody),
		http.StatusOK,
	).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// When
	err := subject.Save(response)

	// Then
	assert.Nil(t, err)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_Read_HappyPath(t *testing.T) {
	// Given
	request := aDummyRequest()
	db, mock, closeFunc := aDatabaseMock(t)
	defer closeFunc()

	subject := &sqliteResponseStore{database: db}

	mockRows := sqlmock.NewRows([]string{"response_body", "status_code"}).
		AddRow("mock-body", 200).
		AddRow("mock-body", 200)
	mock.ExpectQuery(selectQuery).WithArgs(
		testHost+testPath,
		http.MethodGet,
		[]byte{},
	).WillReturnRows(mockRows)

	// When
	response, err := subject.Read(request)

	// Then
	assert.Nil(t, err)
	assert.NotNil(t, response)
	if err := mock.ExpectationsWereMet(); err != nil {
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
