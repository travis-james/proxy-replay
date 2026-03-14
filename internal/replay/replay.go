package replay

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"

	"github.com/travis-james/proxy-replay/internal/types"
)

var logger = log.New(os.Stdout, "replay: ", log.LstdFlags)

func Replay(store types.Storage, addr string) error {
	handler := ReplayHandler(store)
	logger.Printf("replay server listening on %s\n", addr)
	return http.ListenAndServe(addr, handler)
}

func ReplayHandler(store types.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Proxy-Replay-Key")
		if key == "" {
			http.Error(w, "missing replay key", http.StatusBadRequest)
			return
		}

		rec, err := store.Load(key)
		if err != nil {
			http.Error(w, "recording not found", http.StatusNotFound)
			return
		}

		for k, vals := range rec.Response.Headers {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}

		body, err := base64.StdEncoding.DecodeString(rec.Response.BodyBase64)
		if err != nil {
			logger.Printf("failed to decode body for key %s: %v", key, err)
			http.Error(w, "invalid body encoding", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(rec.Response.StatusCode)
		w.Write(body)
	}
}
