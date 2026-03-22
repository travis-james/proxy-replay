package types

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

type Recording struct {
	Request  RecordedRequest  `json:"request"`
	Response RecordedResponse `json:"response"`
}

/*
	Maybe in the future we'd want to use a DB rather than file to disk, so

going to leave it open as an interface.
*/
type Storage interface {
	Save(key string, rec Recording) (err error)
	Load(key string) (rec Recording, err error)
}
