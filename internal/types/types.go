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

// Recording is the object/type that gets saved to disk for later
// recall.
type Recording struct {
	Request  RecordedRequest  `json:"request"`
	Response RecordedResponse `json:"response"`
}

// Storage interface so it can be implemented with a database, or other
// stores. For this repo, currently only implemented for Files.
type Storage interface {
	Save(key string, rec Recording) (err error)
	Load(key string) (rec Recording, err error)
}
