package recorder

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	logger              = log.New(os.Stdout, "recorder: ", log.LstdFlags)
	sendAndReceiveFunc  = sendAndReceive // Could this causes issues with parallel code, should I even concern myself about that?
	processResponseFunc = processResponse
)

func Record(req RecordedRequest) error {
	rawResp, err := sendAndReceiveFunc(req)
	if err != nil {
		return err
	}

	resp := processResponseFunc(*rawResp)

	got, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	logger.Println("recorded response: ", string(got))
	return nil
}

func sendAndReceive(rr RecordedRequest) (*rawResponse, error) {
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

func processResponse(rawResp rawResponse) RecordedResponse {
	encodedBody := base64.StdEncoding.EncodeToString(rawResp.Body)

	return RecordedResponse{
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
