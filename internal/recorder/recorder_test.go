package recorder

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRecord(t *testing.T) {
	var (
		origSend    = sendAndReceiveFunc
		origProcess = processResponseFunc
	)
	t.Run("success", func(t *testing.T) {
		t.Cleanup(func() {
			sendAndReceiveFunc = origSend
			processResponseFunc = origProcess
		})
		// stub sendAndReceive
		sendAndReceiveFunc = func(rr RecordedRequest) (*rawResponse, error) {
			return &rawResponse{
				StatusCode: 200,
				Body:       []byte("ok"),
				Headers:    map[string][]string{},
			}, nil
		}

		// stub processResponse
		processResponseFunc = func(raw rawResponse) RecordedResponse {
			return RecordedResponse{
				StatusCode: 200,
				BodyBase64: "b2s=", // "ok"
				Headers:    map[string][]string{},
			}
		}

		err := Record(RecordedRequest{})
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("sendAndReceive error", func(t *testing.T) {
		t.Cleanup(func() {
			sendAndReceiveFunc = origSend
			processResponseFunc = origProcess
		})
		sendAndReceiveFunc = func(rr RecordedRequest) (*rawResponse, error) {
			return nil, errors.New("TEST ERROR")
		}

		err := Record(RecordedRequest{})
		if err == nil {
			t.Fatalf("expected error, got nil")
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

	rr := RecordedRequest{
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
		rr := RecordedRequest{
			Method: http.MethodGet, URL: "://bad-url",
		}
		_, err := sendAndReceive(rr)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("unreachable host causes Do() error", func(t *testing.T) {
		rr := RecordedRequest{
			Method: http.MethodGet,
			URL:    "http://127.0.0.1:1",
		}
		_, err := sendAndReceive(rr)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("body read error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.(http.Flusher).Flush()                 // send headers.
			conn, _, _ := w.(http.Hijacker).Hijack() // takeover the socket.
			conn.Close()                             // close connection to force read error
		}))
		defer ts.Close()
		rr := RecordedRequest{Method: http.MethodGet, URL: ts.URL}
		_, err := sendAndReceive(rr)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestProcessResposne(t *testing.T) {
	// set up.
	rr := rawResponse{
		Headers: http.Header{
			"Content-Type":  {"one"},          // should be kept
			"Cache-Control": {"two", "three"}, // should be kept
			"ignored":       {"four"},         // should be dropped
			"skipme":        {"five"},         // should be dropped
		},
		Body:       []byte("hello"),
		StatusCode: 200,
	}
	expectedHeaders := map[string][]string{
		"Content-Type":  {"one"},
		"Cache-Control": {"two", "three"},
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

func TestFilterHeaders(t *testing.T) {
	// set up.
	tests := []struct {
		name            string
		input           http.Header
		headersToRecord map[string]struct{}
		expected        map[string][]string
	}{
		{
			name:     "empty",
			input:    http.Header{},
			expected: map[string][]string{},
		},
		{
			name: "filters only allowed headers",
			input: http.Header{
				"Content-Type":  {"one"},          // should be kept
				"Cache-Control": {"two", "three"}, // should be kept
				"ignored":       {"four"},         // should be dropped
				"skipme":        {"five"},         // should be dropped
			},
			expected: map[string][]string{
				"Content-Type":  {"one"},
				"Cache-Control": {"two", "three"},
			},
		},
	}

	// run and assert.
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := filterHeaders(test.input)

			if !reflect.DeepEqual(got, test.expected) {
				t.Fatalf("expected %v\ngot %v", test.expected, got)
			}
		})
	}
}
