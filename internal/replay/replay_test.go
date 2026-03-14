package replay

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/travis-james/proxy-replay/internal/types"
)

func TestReplayHandler(t *testing.T) {
	var (
		mockStore = &mockStorage{}
		handler   = ReplayHandler(mockStore)
		tests     = []struct {
			key        string
			statusCode int
			headers    string
			body       string
			error      string
		}{
			{
				key:        "",
				statusCode: http.StatusBadRequest,
				headers:    "text/plain; charset=utf-8",
				body:       "missing replay key\n",
			},
			{
				key:        "not-found",
				statusCode: http.StatusNotFound,
				headers:    "text/plain; charset=utf-8",
				body:       "recording not found\n",
			},
			{
				key:        "invalid-body",
				statusCode: http.StatusInternalServerError,
				headers:    "text/plain; charset=utf-8",
				body:       "invalid body encoding\n",
			},
			{
				key:        "happy path",
				statusCode: http.StatusOK,
				headers:    "text/plain",
				body:       "Hello",
			},
		}
	)

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-Proxy-Replay-Key", test.key)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != test.statusCode {
				t.Fatalf("expected status %d, got %d", test.statusCode, rr.Code)
			}

			if ct := rr.Header().Get("Content-Type"); ct != test.headers {
				t.Fatalf("expected: %s, got: %s", test.headers, ct)
			}

			if rr.Body.String() != test.body {
				t.Fatalf("expected body %q, got %q", test.body, rr.Body.String())
			}
		})
	}
}

type mockStorage struct{}

func (m *mockStorage) Load(key string) (types.Recording, error) {
	switch key {
	case "not-found":
		return types.Recording{}, errors.New("record not found")
	case "invalid-body":
		return types.Recording{
			Response: types.RecordedResponse{
				BodyBase64: "invalid body",
			},
		}, nil
	default:
		return types.Recording{
			Response: types.RecordedResponse{
				StatusCode: 200,
				BodyBase64: "SGVsbG8=", // "Hello"
				Headers: map[string][]string{
					"Content-Type": {"text/plain"},
				},
			},
		}, nil
	}
}

func (m *mockStorage) Save(key string, rec types.Recording) error {
	return nil
}

func (m *mockStorage) List() ([]types.RecordingMeta, error) {
	return nil, nil
}
