package replay

import (
	"errors"
	"log"
	"os"

	"github.com/travis-james/proxy-replay/internal/types"
)

var (
	logger          = log.New(os.Stdout, "replay: ", log.LstdFlags)
	ERR_MISSING_KEY = "missing replay key"
)

// Replay the given response named as 'key' in the given 'store.'
func Replay(store types.Storage, key string) (types.RecordedResponse, error) {
	logger.Printf("replay start key=%s", key)

	if key == "" {
		err := errors.New(ERR_MISSING_KEY)
		logger.Printf("replay failed key=%s error=%v", key, err)
		return types.RecordedResponse{}, err
	}

	rec, err := store.Load(key)
	if err != nil {
		logger.Printf("replay load failed key=%s error=%v", key, err)
		return types.RecordedResponse{}, err
	}

	logger.Printf("replay success key=%s status=%d", key, rec.Response.StatusCode)
	return rec.Response, nil
}
