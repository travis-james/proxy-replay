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

var (
	logger     = log.New(os.Stdout, "server: ", log.LstdFlags)
	recordFunc = recorder.Record
	replayFunc = replay.Replay
)

type Server struct {
	store types.Storage
}

func New(store types.Storage) *Server {
	return &Server{
		store: store,
	}
}

// ServeHTTP will execute a given flow for replay/record based on
// header values.
func (s *Server) ServeHTTP(respW http.ResponseWriter, req *http.Request) {
	mode := req.Header.Get("X-Proxy-Replay-Mode")
	key := req.Header.Get("X-Proxy-Replay-Key")

	logger.Printf("incoming request method=%s path=%s mode=%s key=%s",
		req.Method, req.URL.Path, mode, key)

	if key == "" {
		logger.Printf("missing key method=%s path=%s", req.Method, req.URL.Path)
		http.Error(respW, "missing replay key", http.StatusBadRequest)
		return
	}

	switch mode {
	case "record":
		s.handleRecord(key, respW, req)
	case "replay":
		s.handleReplay(key, respW, req)
	default:
		logger.Printf("invalid mode key=%s mode=%s", key, mode)
		http.Error(respW, "invalid mode", http.StatusBadRequest)
	}
}

// handleReplay takes a given recorded response based on 'key' and writes
// the result to a ResponseWriter.
func (s *Server) handleReplay(key string, respW http.ResponseWriter, _ *http.Request) {
	logger.Printf("replay request key=%s", key)

	resp, err := replayFunc(s.store, key)
	if err != nil {
		logger.Printf("replay failed key=%s error=%v", key, err)
		http.Error(respW, "recording not found", http.StatusNotFound)
		return
	}

	for k, vals := range resp.Headers {
		for _, v := range vals {
			respW.Header().Add(k, v)
		}
	}

	body, err := base64.StdEncoding.DecodeString(resp.BodyBase64)
	if err != nil {
		logger.Printf("replay decode failed key=%s error=%v", key, err)
		http.Error(respW, "invalid body encoding", http.StatusInternalServerError)
		return
	}

	respW.WriteHeader(resp.StatusCode)
	respW.Write(body)

	logger.Printf("replay success key=%s status=%d", key, resp.StatusCode)
}

// handleRecord takes a given
func (s *Server) handleRecord(key string, respW http.ResponseWriter, req *http.Request) {
	logger.Printf("record request key=%s", key)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Printf("failed to read request body key=%s error=%v", key, err)
		http.Error(respW, "failed to read body", 500)
		return
	}

	recReq := types.RecordedRequest{
		Method:  req.Method,
		URL:     req.Header.Get("X-Proxy-Target-URL"),
		Headers: req.Header,
		Body:    body,
	}

	rawResp, err := recordFunc(s.store, key, recReq)
	if err != nil {
		logger.Printf("record failed key=%s error=%v", key, err)
		http.Error(respW, err.Error(), 500)
		return
	}

	// return REAL response to client
	for k, vals := range rawResp.Headers {
		for _, v := range vals {
			respW.Header().Add(k, v)
		}
	}

	respW.WriteHeader(rawResp.StatusCode)
	respW.Write(rawResp.Body)

	logger.Printf("record success key=%s status=%d", key, rawResp.StatusCode)
}
