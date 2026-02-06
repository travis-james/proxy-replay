package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/travis-james/proxy-replay/internal/recorder"
)

func TestFileStorage_Save(t *testing.T) {
	// set up.
	var (
		dir      = t.TempDir()
		fs       = FileStorage{Dir: dir}
		fileName = "test-recording.json"
		req      = recorder.RecordedRequest{
			Method: "GET",
			URL:    "https://example.com",
			Headers: map[string][]string{
				"Accept": {"application/json"},
			},
			Body: []byte("hello"),
		}
		resp = recorder.RecordedResponse{
			StatusCode: 200,
			BodyBase64: "aGVsbG8=",
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
		}
	)

	// run.
	if err := fs.Save(fileName, req, resp); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	// assert.
	// get the file that was saved.
	finalPath := filepath.Join(dir, fileName)
	data, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	// unmarshal and verify contents
	var rec Recording
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if rec.Request.Method != req.Method {
		t.Errorf("expected method %q, got %q", req.Method, rec.Request.Method)
	}
	if rec.Response.StatusCode != resp.StatusCode {
		t.Errorf("expected status %d, got %d", resp.StatusCode, rec.Response.StatusCode)
	}
}
