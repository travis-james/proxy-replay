package storage

import (
	"time"

	"github.com/travis-james/proxy-replay/internal/recorder"
)

type Storage interface {
	Save(req recorder.RecordedRequest, resp recorder.RecordedResponse) (key string, err error)
	Load(key string) (recorder.RecordedRequest, recorder.RecordedResponse, error)
	List() ([]RecordingMeta, error)
}

type RecordingMeta struct {
	Key       string // filename/request name
	Method    string
	URL       string
	Timestamp time.Time // when it was recorded
	SizeBytes int64     // size of the stored file
}
