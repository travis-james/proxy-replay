package recorder

import (
	"bytes"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/travis-james/proxy-replay/internal/types"
)

var (
	logger             = log.New(os.Stdout, "recorder: ", log.LstdFlags)
	sendAndReceiveFunc = sendAndReceive // Could this causes issues with parallel code, should I even concern myself about that?
)

// rawResponse is an intermediary data structure
// from what we receive from the destination URL/source
// to what eventually becomes a new structure, recorded
// response.
type rawResponse struct {
	Headers    http.Header
	Body       []byte
	StatusCode int
}

func Record(store types.Storage, key string, req types.RecordedRequest) (*rawResponse, error) {
	logger.Printf("record start key=%s method=%s url=%s", key, req.Method, req.URL)
	rawResp, err := sendAndReceiveFunc(req)
	if err != nil {
		logger.Printf("record failed key=%s error=%v", key, err)
		return nil, err
	}

	record := types.Recording{
		Request:  req,
		Response: processResponse(*rawResp),
	}

	if err := store.Save(key, record); err != nil {
		logger.Printf("record save failed key=%s error=%v", key, err)
		return nil, err
	}

	logger.Printf("record success key=%s status=%d", key, rawResp.StatusCode)
	return rawResp, nil
}

func sendAndReceive(rr types.RecordedRequest) (*rawResponse, error) {
	logger.Printf("outbound request method=%s url=%s", rr.Method, rr.URL)

	// Build request.
	req, err := http.NewRequest(rr.Method, rr.URL, bytes.NewReader(rr.Body))
	if err != nil {
		logger.Printf("failed to build request url=%s error=%v", rr.URL, err)
		return nil, err
	}

	// add headers.
	for k, vals := range rr.Headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}

	// send
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Printf("request failed url=%s error=%v", rr.URL, err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("failed to read response body url=%s error=%v", rr.URL, err)
		return nil, err
	}

	logger.Printf("response received status=%d bytes=%d url=%s",
		resp.StatusCode, len(body), rr.URL)

	return &rawResponse{
		Headers:    resp.Header,
		Body:       body,
		StatusCode: resp.StatusCode,
	}, nil
}

func processResponse(rawResp rawResponse) types.RecordedResponse {
	encodedBody := base64.StdEncoding.EncodeToString(rawResp.Body)

	headersCopy := make(map[string][]string)
	for k, vals := range rawResp.Headers {
		copiedVals := make([]string, len(vals))
		copy(copiedVals, vals)
		headersCopy[k] = copiedVals
	}

	return types.RecordedResponse{
		StatusCode: rawResp.StatusCode,
		BodyBase64: encodedBody,
		Headers:    headersCopy,
	}
}
