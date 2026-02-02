package recorder

import (
	"net/http"
	"reflect"
	"testing"
)

func TestProcessResposne(t *testing.T) {

}

func TestFilterHeaders(t *testing.T) {
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := filterHeaders(test.input)

			if !reflect.DeepEqual(got, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, got)
			}
		})
	}
}
