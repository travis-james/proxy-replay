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
	// HeadersToRecord acts as a set to COMPARE against
	// for what headers we want to keep/record on the
	// received response.
	HeadersToRecord = map[string]struct{}{ // A set.
		"Content-Type":  {},
		"Date":          {},
		"Server":        {},
		"Cache-Control": {},
	}
	logger              = log.New(os.Stdout, "recorder: ", log.LstdFlags)
	sendAndReceiveFunc  = sendAndReceive // Could this causes issues with parallel code, should I even concern myself about that?
	processResponseFunc = processResponse
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

func Record(store types.Storage, key string, req types.RecordedRequest) error {
	rawResp, err := sendAndReceiveFunc(req)
	if err != nil {
		return err
	}

	record := types.Recording{
		Request:  req,
		Response: processResponseFunc(*rawResp),
	}

	if err := store.Save(key, record); err != nil {
		return err
	}

	logger.Println("recorded request/response with key: ", key)
	return nil
}

func sendAndReceive(rr types.RecordedRequest) (*rawResponse, error) {
	logger.Println("sending request to: ", rr.URL)

	// Build request.
	req, err := http.NewRequest(rr.Method, rr.URL, bytes.NewReader(rr.Body))
	if err != nil {
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
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	logger.Printf("resp status code: %d, resp body length: %d\n", resp.StatusCode, len(body))

	return &rawResponse{
		Headers:    resp.Header,
		Body:       body,
		StatusCode: resp.StatusCode,
	}, nil
}

func processResponse(rawResp rawResponse) types.RecordedResponse {
	encodedBody := base64.StdEncoding.EncodeToString(rawResp.Body)

	return types.RecordedResponse{
		StatusCode: rawResp.StatusCode,
		BodyBase64: encodedBody,
		Headers:    filterHeaders(rawResp.Headers),
	}
}

func filterHeaders(headers http.Header) map[string][]string {
	retval := map[string][]string{}
	for key, val := range headers {
		if _, ok := HeadersToRecord[key]; !ok {
			continue
		}
		retval[key] = val
	}
	return retval
}
