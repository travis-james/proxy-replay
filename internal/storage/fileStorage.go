package storage

import "github.com/travis-james/proxy-replay/internal/recorder"

type FileStorage struct {
	Dir string
}

func (fs FileStorage) Save(req recorder.RecordedRequest, resp recorder.RecordedResponse) (key string, err error) {
	return
}

func (fs FileStorage) Load(key string) (req recorder.RecordedRequest, resp recorder.RecordedResponse, err error) {
	return
}

func (fs FileStorage) List() (metaData []RecordingMeta, err error) {
	return
}
