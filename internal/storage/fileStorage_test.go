package storage

import (
	"reflect"
	"testing"

	"github.com/travis-james/proxy-replay/internal/recorder"
)

func TestFileStorage_SaveAndLoad(t *testing.T) {
	// set up.
	var (
		dir      = t.TempDir()
		fs       = FileStorage{Dir: dir}
		fileName = "test-recording"
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
	err := fs.Save(fileName, rec)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := fs.Load(fileName)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !reflect.DeepEqual(rec, loaded) {
		t.Fatalf("recording mismatch\nexpected: %+v\ngot: %+v", rec, loaded)
	}
}
