package recorder

import "net/http"

var HeadersToRecord = map[string]struct{}{ // A set.
	"Content-Type":  {},
	"Date":          {},
	"Server":        {},
	"Cache-Control": {},
}

type RawResponse struct {
	Headers    http.Header
	Body       []byte
	StatusCode int
}

type RecordedResponse struct {
	StatusCode int
	BodyBase64 string
	Headers    map[string][]string
}

type RecordedRequest struct {
	Method  string
	URL     string
	Headers map[string][]string
	Body    []byte
}
