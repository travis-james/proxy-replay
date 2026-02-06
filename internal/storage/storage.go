package storage

import (
	"time"

	"github.com/travis-james/proxy-replay/internal/recorder"
)

/* Maybe in the future we'd want to use a DB rather than file to disk, so
going to leave it open as an interface. */

type Storage interface {
	Save(req recorder.RecordedRequest, resp recorder.RecordedResponse) (key string, err error)
	Load(key string) (req recorder.RecordedRequest, resp recorder.RecordedResponse, err error)
	List() (metaData []RecordingMeta, err error)
}

type RecordingMeta struct {
	Key       string // filename/request name
	Method    string
	URL       string
	Timestamp time.Time // when it was recorded
	SizeBytes int64     // size of the stored file
}
