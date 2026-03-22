package replay

import (
	"errors"
	"testing"

	"github.com/travis-james/proxy-replay/internal/types"
)

func TestReplay(t *testing.T) {
	mockStore := &mockStorage{}

	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "missing key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "not found",
			key:     "not-found",
			wantErr: true,
		},
		{
			name: "success",
			key:  "happy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := Replay(mockStore, tt.key)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.StatusCode != 200 {
				t.Fatalf("expected status 200, got %d", resp.StatusCode)
			}

			if resp.BodyBase64 != "SGVsbG8=" {
				t.Fatalf("unexpected body: %s", resp.BodyBase64)
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
