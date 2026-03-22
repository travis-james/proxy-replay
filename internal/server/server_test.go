package server

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/travis-james/proxy-replay/internal/storage"
	"github.com/travis-james/proxy-replay/internal/types"
)

func TestHandleRecord(t *testing.T) {
	old := recordFunc
	defer func() { recordFunc = old }()
	tests := []struct {
		testName       string
		respRec        *httptest.ResponseRecorder
		req            *http.Request
		mockRecordFunc func(store types.Storage, key string, req types.RecordedRequest) (*types.RawResponse, error)
		expectedCode   int
		expectedBody   string
		expectedHeader string
	}{
		{
			testName: "happy path",
			respRec:  httptest.NewRecorder(),
			req:      httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("hello"))),
			mockRecordFunc: func(store types.Storage, key string, req types.RecordedRequest) (*types.RawResponse, error) {
				return &types.RawResponse{
					StatusCode: 201,
					Body:       []byte("recorded"),
					Headers:    http.Header{"X-Test": []string{"ok"}},
				}, nil
			},
			expectedCode:   201,
			expectedBody:   "recorded",
			expectedHeader: "ok",
		},
		{
			testName: "record failure",
			respRec:  httptest.NewRecorder(),
			req:      httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("mock error"))),
			mockRecordFunc: func(store types.Storage, key string, req types.RecordedRequest) (*types.RawResponse, error) {
				return &types.RawResponse{
					StatusCode: 500,
					Body:       []byte("ohno"),
				}, errors.New("mock error")
			},
			expectedCode: 500,
			expectedBody: "mock error\n",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			recordFunc = test.mockRecordFunc

			test.req.Header.Set("X-Proxy-Target-URL", "http://example.com")

			serv := New(storage.FileStorage{})
			serv.handleRecord("key", test.respRec, test.req)
			// assertions
			if test.respRec.Code != test.expectedCode {
				t.Fatalf("expected %d, got %d", test.expectedCode, test.respRec.Code)
			}

			if test.respRec.Body.String() != test.expectedBody {
				t.Fatalf("expected body %q, got %q", test.expectedBody, test.respRec.Body.String())
			}

			if test.respRec.Header().Get("X-Test") != test.expectedHeader {
				t.Fatalf("expected header %s, got %s", test.expectedHeader, test.respRec.Header().Get("X-Test"))
			}
		})
	}
}
