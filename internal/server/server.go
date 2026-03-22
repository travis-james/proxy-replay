package server

import (
	"encoding/base64"
	"io"
	"net/http"

	"github.com/travis-james/proxy-replay/internal/recorder"
	"github.com/travis-james/proxy-replay/internal/replay"
	"github.com/travis-james/proxy-replay/internal/types"
)

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
	if key == "" {
		http.Error(w, "invalid mode", http.StatusBadRequest)
		return
	}

	switch mode {
	case "record":
		s.handleRecord(key, w, r)
	case "replay":
		s.handleReplay(key, w, r)
	default:
		http.Error(w, "invalid mode", http.StatusBadRequest)
	}
}

func (s *Server) handleReplay(key string, w http.ResponseWriter, r *http.Request) {

	resp, err := replay.Replay(s.store, key)
	if err != nil {
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
		http.Error(w, "invalid body encoding", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func (s *Server) handleRecord(key string, w http.ResponseWriter, r *http.Request) {

	body, _ := io.ReadAll(r.Body)

	req := types.RecordedRequest{
		Method:  r.Method,
		URL:     r.Header.Get("X-Proxy-Target-URL"),
		Headers: r.Header,
		Body:    body,
	}

	rawResp, err := recorder.Record(s.store, key, req)
	if err != nil {
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
}
