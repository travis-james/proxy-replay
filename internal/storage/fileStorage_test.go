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
		rec = Recording{
			Request:  req,
			Response: resp,
		}
	)

	// run.
	if err := fs.Save(fileName, rec); err != nil {
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
	var got Recording
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Request.Method != req.Method {
		t.Errorf("expected method %q, got %q", req.Method, got.Request.Method)
	}
	if got.Response.StatusCode != resp.StatusCode {
		t.Errorf("expected status %d, got %d", resp.StatusCode, got.Response.StatusCode)
	}
}
