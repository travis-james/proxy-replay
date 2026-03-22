package replay

import (
	"errors"
	"testing"

	"github.com/travis-james/proxy-replay/internal/types"
)

var (
	missingKeyError  = "missing replay key"
	noRecordingError = "no record for the given key"
)

func TestReplay(t *testing.T) {
	mockStore := &mockStorage{}

	tests := []struct {
		name               string
		key                string
		expectedError      string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:          "missing key",
			key:           "",
			expectedError: missingKeyError,
		},
		{
			name:          "not found",
			key:           "not-found",
			expectedError: noRecordingError,
		},
		{
			name:               "success",
			key:                "happy",
			expectedStatusCode: 200,
			expectedBody:       "SGVsbG8=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := Replay(mockStore, tt.key)
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != tt.expectedError {
				t.Fatalf("expected error %q, got %q", tt.expectedError, errMsg)
			}

			if resp.StatusCode != tt.expectedStatusCode {
				t.Fatalf("expected status 200, got %d", resp.StatusCode)
			}

			if resp.BodyBase64 != tt.expectedBody {
				t.Fatalf("unexpected body: %s", resp.BodyBase64)
			}
		})
	}
}

type mockStorage struct{}

func (m *mockStorage) Load(key string) (types.Recording, error) {
	if key == "not-found" {
		return types.Recording{}, errors.New(noRecordingError)
	}

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

func (m *mockStorage) Save(key string, rec types.Recording) error {
	return nil
}
