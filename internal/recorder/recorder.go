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

// Record the given request and eventual response to the given store
// with the name of key.
func Record(store types.Storage, key string, req types.RecordedRequest) (types.RawResponse, error) {
	logger.Printf("record start key=%s method=%s url=%s", key, req.Method, req.URL)

	rawResp, err := sendAndReceiveFunc(req)
	if err != nil {
		logger.Printf("record failed key=%s error=%v", key, err)
		return types.RawResponse{}, err
	}

	record := types.Recording{
		Request:  req,
		Response: processResponse(rawResp),
	}

	if err := store.Save(key, record); err != nil {
		logger.Printf("record save failed key=%s error=%v", key, err)
		return types.RawResponse{}, err
	}

	logger.Printf("record success key=%s status=%d", key, rawResp.StatusCode)
	return rawResp, nil
}

// sendAndReceive takes the input request and makes an HTTP call with
// it returning the response from that call.
func sendAndReceive(rr types.RecordedRequest) (types.RawResponse, error) {
	logger.Printf("outbound request method=%s url=%s", rr.Method, rr.URL)

	// Build request.
	req, err := http.NewRequest(rr.Method, rr.URL, bytes.NewReader(rr.Body))
	if err != nil {
		logger.Printf("failed to build request url=%s error=%v", rr.URL, err)
		return types.RawResponse{}, err
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
		return types.RawResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("failed to read response body url=%s error=%v", rr.URL, err)
		return types.RawResponse{}, err
	}

	logger.Printf("response received status=%d bytes=%d url=%s",
		resp.StatusCode, len(body), rr.URL)

	return types.RawResponse{
		Headers:    resp.Header,
		Body:       body,
		StatusCode: resp.StatusCode,
	}, nil
}

// processResponse is a helper to transform the rawResponse to the
// RecordedResponse that gets saved to storage.
func processResponse(rawResp types.RawResponse) types.RecordedResponse {
	// Base64-encode body because raw response bytes may not be
	// valid UTF-8 and cannot be safely embedded in JSON format
	// that is used when saved to disk.
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
