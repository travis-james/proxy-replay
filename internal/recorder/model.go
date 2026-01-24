package recorder

import (
	"encoding/base64"
	"fmt"
)

type RecordedResponse struct {
	StatusCode int
	BodyBase64 string
	Headers    map[string][]string
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
