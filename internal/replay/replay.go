package replay

import (
	"errors"

	"github.com/travis-james/proxy-replay/internal/types"
)

func Replay(store types.Storage, key string) (types.RecordedResponse, error) {
	if key == "" {
		return types.RecordedResponse{}, errors.New("missing replay key")
	}

	rec, err := store.Load(key)
	if err != nil {
		return types.RecordedResponse{}, err
	}

	return rec.Response, nil
}
