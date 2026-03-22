package recorder

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/travis-james/proxy-replay/internal/types"
)

var testStoreageError = "TEST STORAGE ERROR"

func TestRecord(t *testing.T) {
	origSend := sendAndReceiveFunc
	t.Cleanup(func() {
		sendAndReceiveFunc = origSend
	})

	t.Run("success", func(t *testing.T) {
		sendAndReceiveFunc = func(rr types.RecordedRequest) (*types.RawResponse, error) {
			return &types.RawResponse{
				StatusCode: 200,
				Body:       []byte("ok"),
				Headers: http.Header{
					"Content-Type": {"text/plain"},
				},
			}, nil
		}

		mock := &mockStorage{}
		key := "test-key"

		resp, err := Record(mock, key, types.RecordedRequest{})
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if resp == nil {
			t.Fatalf("expected response, got nil")
		}
		if resp.StatusCode != 200 {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		if mock.savedKey != key {
			t.Fatalf("expected %s, got %s", key, mock.savedKey)
		}

		if mock.savedRec.Response.StatusCode != 200 {
			t.Fatalf("expected saved status 200, got %d", mock.savedRec.Response.StatusCode)
		}

		// base64("ok") = "b2s="
		if mock.savedRec.Response.BodyBase64 != "b2s=" {
			t.Fatalf("unexpected encoded body: %s", mock.savedRec.Response.BodyBase64)
		}

		if ct := mock.savedRec.Response.Headers["Content-Type"][0]; ct != "text/plain" {
			t.Fatalf("expected Content-Type text/plain, got %s", ct)
		}
	})

	t.Run("sendAndReceive error", func(t *testing.T) {
		sendAndReceiveFunc = func(rr types.RecordedRequest) (*types.RawResponse, error) {
			return nil, errors.New("TEST ERROR")
		}

		mock := &mockStorage{}
		key := "test-key"

		resp, err := Record(mock, key, types.RecordedRequest{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if resp != nil {
			t.Fatalf("expected nil response, got %+v", resp)
		}
	})

	t.Run("storage save error", func(t *testing.T) {
		sendAndReceiveFunc = func(rr types.RecordedRequest) (*types.RawResponse, error) {
			return &types.RawResponse{
				StatusCode: 200,
				Body:       []byte("ok"),
				Headers:    http.Header{},
			}, nil
		}

		mock := &mockStorage{}
		key := "fail" // triggers error

		resp, err := Record(mock, key, types.RecordedRequest{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != testStoreageError {
			t.Fatalf("expected %v, got %v", testStoreageError, err.Error())
		}
		if resp != nil {
			t.Fatalf("expected nil response, got %+v", resp)
		}
	})
}

func TestSendAndReceive(t *testing.T) {
	// set up.
	var (
		statusCode = 201
		body       = `{"ok":true}`
		headerKey  = "Content-Type"
		headerVal  = "application/json"
		xtestVal   = "TestSendAndReceive"
	)
	// The 'external' server who's response we want to record.
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if val := r.Header.Get("X-Test"); val != xtestVal {
			t.Fatalf("expected header val to be %s, got %s", xtestVal, val)
		}

		w.Header().Set(headerKey, headerVal)
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}))
	defer testServer.Close()

	rr := types.RecordedRequest{
		Method:  http.MethodPost,
		URL:     testServer.URL,
		Headers: map[string][]string{"X-Test": {xtestVal}},
		Body:    []byte("hello"),
	}

	// run.
	got, err := sendAndReceive(rr)

	// assert.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got.Body) != body {
		t.Fatalf("body mismatch\nexpected: '%s', got '%s'", body, string(got.Body))
	}
	if val := got.Headers.Get(headerKey); val != headerVal {
		t.Fatalf("header mismatch\nexpected: '%s', got: '%s'", headerVal, val)
	}
}

func TestSendAndReceiveErrors(t *testing.T) {
	t.Run("invalid URL causes NewRequest error", func(t *testing.T) {
		expectedErr := `parse "://bad-url": missing protocol scheme`
		rr := types.RecordedRequest{
			Method: http.MethodGet, URL: "://bad-url",
		}
		_, err := sendAndReceive(rr)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != expectedErr {
			t.Fatalf("expected: %v, got: %v", expectedErr, err.Error())
		}
	})

	t.Run("unreachable host causes Do() error", func(t *testing.T) {
		expectedErr := `Get "http://127.0.0.1:1": dial tcp 127.0.0.1:1: connect: connection refused`
		rr := types.RecordedRequest{
			Method: http.MethodGet,
			URL:    "http://127.0.0.1:1",
		}
		_, err := sendAndReceive(rr)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != expectedErr {
			t.Fatalf("expected: %v, got: %v", expectedErr, err.Error())
		}
	})

	t.Run("body read error", func(t *testing.T) {
		expectedErr := "unexpected EOF"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.(http.Flusher).Flush()                 // send headers immediately so client sees the start of an http response...
			conn, _, _ := w.(http.Hijacker).Hijack() // takeover the socket.
			conn.Close()                             // close connection to force read error, client doesn't get the Body.
		}))
		defer ts.Close()
		rr := types.RecordedRequest{Method: http.MethodGet, URL: ts.URL}
		_, err := sendAndReceive(rr)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if err.Error() != expectedErr {
			t.Fatalf("expected: %v, got: %v", expectedErr, err.Error())
		}

	})
}

func TestProcessResposne(t *testing.T) {
	// set up.
	rr := types.RawResponse{
		Headers: http.Header{
			"Content-Type":  {"one"},
			"Cache-Control": {"two", "three"},
			"X-Custom":      {"four"},
		},
		Body:       []byte("hello"),
		StatusCode: 200,
	}
	expectedHeaders := map[string][]string{
		"Content-Type":  {"one"},
		"Cache-Control": {"two", "three"},
		"X-Custom":      {"four"},
	}
	expectedBody := "aGVsbG8="

	// run.
	got := processResponse(rr)

	// assert.
	switch {
	case !reflect.DeepEqual(expectedHeaders, got.Headers):
		t.Fatalf("header mismatch\nexpected %v\ngot %v", expectedHeaders, got.Headers)
	case expectedBody != got.BodyBase64:
		t.Fatalf("body mismatch\nexpected %v\ngot %v", expectedBody, got.BodyBase64)
	case rr.StatusCode != got.StatusCode:
		t.Fatalf("status code mismatch\nexpected %v\ngot %v", rr.StatusCode, got.StatusCode)
	}
}

type mockStorage struct {
	savedKey string
	savedRec types.Recording
}

func (m *mockStorage) Save(key string, rec types.Recording) error {
	if key == "fail" {
		return errors.New(testStoreageError)
	}
	m.savedKey = key
	m.savedRec = rec
	return nil
}

func (m *mockStorage) Load(key string) (types.Recording, error) {
	return types.Recording{}, nil
}
