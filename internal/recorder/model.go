package recorder

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

var HeadersToRecord = map[string]struct{}{ // A set.
	"Content-Type":  {},
	"Date":          {},
	"Server":        {},
	"Cache-Control": {},
}

type RecordedResponse struct {
	StatusCode int
	BodyBase64 string
	Headers    map[string][]string
}

func (rr RecordedResponse) ToJSON() ([]byte, error) {
	jsonData, err := json.Marshal(rr)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (rr RecordedResponse) Stringify(decode bool) (string, error) {
	if decode {
		body, err := base64.StdEncoding.DecodeString(rr.BodyBase64)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(
			"StatusCode: %d\nHeaders: %v\nBody: %s",
			rr.StatusCode, rr.Headers, body), nil
	}
	return fmt.Sprintf(
		"StatusCode: %d\nHeaders: %v\nBodyBase64: %s",
		rr.StatusCode, rr.Headers, rr.BodyBase64), nil
}

type RecordedRequest struct {
	Method     string
	URL        string
	Headers    map[string][]string
	BodyBase64 string
}
