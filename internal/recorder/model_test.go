package recorder

import "testing"

func TestToJSON(t *testing.T) {
	teststruct := RecordedResponse{
		StatusCode: 200,
		BodyBase64: "hello",
		Headers: map[string][]string{
			"test": {"1", "2", "3"},
			"foo":  {"4", "5", "6"},
		},
	}
	got, err := teststruct.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal(string(got))
}
