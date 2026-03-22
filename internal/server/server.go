package server

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/travis-james/proxy-replay/internal/recorder"
	"github.com/travis-james/proxy-replay/internal/replay"
	"github.com/travis-james/proxy-replay/internal/types"
)

var logger = log.New(os.Stdout, "server: ", log.LstdFlags)

type Server struct {
	store types.Storage
}

func New(store types.Storage) *Server {
	return &Server{
		store: store,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mode := r.Header.Get("X-Proxy-Replay-Mode")
	key := r.Header.Get("X-Proxy-Replay-Key")

	logger.Printf("incoming request method=%s path=%s mode=%s key=%s",
		r.Method, r.URL.Path, mode, key)

	if key == "" {
		logger.Printf("missing key method=%s path=%s", r.Method, r.URL.Path)
		http.Error(w, "missing replay key", http.StatusBadRequest)
		return
	}

	switch mode {
	case "record":
		s.handleRecord(key, w, r)
	case "replay":
		s.handleReplay(key, w, r)
	default:
		logger.Printf("invalid mode key=%s mode=%s", key, mode)
		http.Error(w, "invalid mode", http.StatusBadRequest)
	}
}

func (s *Server) handleReplay(key string, w http.ResponseWriter, _ *http.Request) {
	logger.Printf("replay request key=%s", key)

	resp, err := replay.Replay(s.store, key)
	if err != nil {
		logger.Printf("replay failed key=%s error=%v", key, err)
		http.Error(w, "recording not found", http.StatusNotFound)
		return
	}

	for k, vals := range resp.Headers {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}

	body, err := base64.StdEncoding.DecodeString(resp.BodyBase64)
	if err != nil {
		logger.Printf("replay decode failed key=%s error=%v", key, err)
		http.Error(w, "invalid body encoding", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	logger.Printf("replay success key=%s status=%d", key, resp.StatusCode)
}

func (s *Server) handleRecord(key string, w http.ResponseWriter, r *http.Request) {
	logger.Printf("replay request key=%s", key)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Printf("failed to read request body key=%s error=%v", key, err)
		http.Error(w, "failed to read body", 500)
		return
	}

	req := types.RecordedRequest{
		Method:  r.Method,
		URL:     r.Header.Get("X-Proxy-Target-URL"),
		Headers: r.Header,
		Body:    body,
	}

	rawResp, err := recorder.Record(s.store, key, req)
	if err != nil {
		logger.Printf("record failed key=%s error=%v", key, err)
		http.Error(w, err.Error(), 500)
		return
	}

	// return REAL response to client
	for k, vals := range rawResp.Headers {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(rawResp.StatusCode)
	w.Write(rawResp.Body)

	logger.Printf("record success key=%s status=%d", key, rawResp.StatusCode)
}
