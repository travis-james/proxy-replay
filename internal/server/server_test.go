package server

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/travis-james/proxy-replay/internal/storage"
	"github.com/travis-james/proxy-replay/internal/types"
)

func TestHandleReplay(t *testing.T) {
	oldReplay := replayFunc
	defer func() { replayFunc = oldReplay }()

	tests := []struct {
		testName          string
		mockReplayFunc    func(store types.Storage, key string) (types.RecordedResponse, error)
		key               string
		expectedStatus    int
		expectedBody      string
		expectedHeaderKey string
		expectedHeaderVal string
	}{
		{
			testName: "happy path",
			mockReplayFunc: func(store types.Storage, key string) (types.RecordedResponse, error) {
				return types.RecordedResponse{
					StatusCode: 200,
					BodyBase64: base64.StdEncoding.EncodeToString([]byte("hello")),
					Headers:    map[string][]string{"X-Test": {"ok"}},
				}, nil
			},
			key:               "k1",
			expectedStatus:    200,
			expectedBody:      "hello",
			expectedHeaderKey: "X-Test",
			expectedHeaderVal: "ok",
		},
		{
			testName: "recording not found",
			mockReplayFunc: func(store types.Storage, key string) (types.RecordedResponse, error) {
				return types.RecordedResponse{}, errors.New("recording not found")
			},
			key:            "k2",
			expectedStatus: 404,
			expectedBody:   "recording not found\n",
		},
		{
			testName: "invalid base64 body",
			mockReplayFunc: func(store types.Storage, key string) (types.RecordedResponse, error) {
				return types.RecordedResponse{
					StatusCode: 200,
					BodyBase64: "notbase64",
				}, nil
			},
			key:            "k3",
			expectedStatus: 500,
			expectedBody:   "invalid body encoding\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			replayFunc = tt.mockReplayFunc

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			serv := New(storage.FileStorage{})
			serv.handleReplay(tt.key, rr, req)

			if rr.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Fatalf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}

			if tt.expectedHeaderKey != "" {
				got := rr.Header().Get(tt.expectedHeaderKey)
				if got != tt.expectedHeaderVal {
					t.Fatalf("expected header %s=%s, got %s", tt.expectedHeaderKey, tt.expectedHeaderVal, got)
				}
			}
		})
	}
}

func TestHandleRecord(t *testing.T) {
	old := recordFunc
	defer func() { recordFunc = old }()
	tests := []struct {
		testName       string
		respRec        *httptest.ResponseRecorder
		req            *http.Request
		mockRecordFunc func(store types.Storage, key string, req types.RecordedRequest) (types.RawResponse, error)
		expectedCode   int
		expectedBody   string
		expectedHeader string
	}{
		{
			testName: "happy path",
			respRec:  httptest.NewRecorder(),
			req:      httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("hello"))),
			mockRecordFunc: func(store types.Storage, key string, req types.RecordedRequest) (types.RawResponse, error) {
				return types.RawResponse{
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
			mockRecordFunc: func(store types.Storage, key string, req types.RecordedRequest) (types.RawResponse, error) {
				return types.RawResponse{
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

func TestServeHTTP(t *testing.T) {
	oldRecord := recordFunc
	recordFunc = func(store types.Storage, key string, req types.RecordedRequest) (types.RawResponse, error) {
		return types.RawResponse{
			StatusCode: 201,
			Body:       []byte("from record"),
			Headers:    http.Header{"X-Test": []string{"ok"}},
		}, nil
	}
	defer func() { recordFunc = oldRecord }()

	oldReplay := replayFunc
	replayFunc = func(store types.Storage, key string) (types.RecordedResponse, error) {
		return types.RecordedResponse{
			StatusCode: 200,
			BodyBase64: base64.StdEncoding.EncodeToString([]byte("from replay")),
			Headers:    map[string][]string{"X-Test": {"ok"}},
		}, nil
	}
	defer func() { replayFunc = oldReplay }()
	server := New(storage.FileStorage{})
	tests := []struct {
		testName     string
		reqBody      []byte
		headers      map[string]string
		expectedCode int
		expectedBody string
	}{
		{
			testName:     "missing key",
			expectedCode: 400,
			expectedBody: "missing replay key\n",
		},
		{
			testName:     "no mode",
			headers:      map[string]string{"X-Proxy-Replay-Key": "test"},
			expectedCode: 400,
			expectedBody: "invalid mode\n",
		},
		{
			testName: "record mode",
			headers: map[string]string{
				"X-Proxy-Replay-Key":  "test",
				"X-Proxy-Replay-Mode": "record",
			},
			expectedCode: 201,
			expectedBody: "from record",
		},
		{
			testName: "replay mode",
			headers: map[string]string{
				"X-Proxy-Replay-Key":  "test",
				"X-Proxy-Replay-Mode": "replay",
			},
			expectedCode: 200,
			expectedBody: "from replay",
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			respRec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(test.reqBody))

			// apply headers
			for k, v := range test.headers {
				req.Header.Set(k, v)
			}

			server.ServeHTTP(respRec, req)

			if respRec.Code != test.expectedCode {
				t.Fatalf("expected %d, got %d", test.expectedCode, respRec.Code)
			}

			if respRec.Body.String() != test.expectedBody {
				t.Fatalf("expected body %q, got %q", test.expectedBody, respRec.Body.String())
			}
		})
	}
}
