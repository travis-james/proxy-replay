package recorder

import "net/http"

// HeadersToRecord acts as a set to compare against
// for what headers we want to keep/record on the
// received response.
var HeadersToRecord = map[string]struct{}{ // A set.
	"Content-Type":  {},
	"Date":          {},
	"Server":        {},
	"Cache-Control": {},
}

// rawResponse is an intermediary data structure
// from what we receive from the destination URL/source
// to what eventually becomes a new structure, recorded
// response.
type rawResponse struct {
	Headers    http.Header
	Body       []byte
	StatusCode int
}

// RecordedResponse is the response from the remote/destination
// server that will then be used for mocking/testing.
type RecordedResponse struct {
	StatusCode int
	BodyBase64 string
	Headers    map[string][]string
}

// RecordedRequest is the request sent from the client to
// this proxy application.
type RecordedRequest struct {
	Method  string
	URL     string
	Headers map[string][]string
	Body    []byte
}
