package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

type RecordedResponse struct {
	StatusCode int
	BodyBase64 string
}

func main() {
	url := "https://jsonplaceholder.typicode.com/todos/1"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	encodedBody := base64.StdEncoding.EncodeToString(body)

	rr := RecordedResponse{
		StatusCode: resp.StatusCode,
		BodyBase64: encodedBody,
	}

	fmt.Println(rr.Stringify(true))
}

func (rr RecordedResponse) Stringify(decode bool) string {
	if decode {
		body, err := base64.StdEncoding.DecodeString(rr.BodyBase64)
		if err != nil {
			panic(err)
		}
		return fmt.Sprintf(
			"StatusCode: %d\nBody: %s",
			rr.StatusCode, body)
	}
	return fmt.Sprintf(
		"StatusCode: %d\nBodyBase64: %s",
		rr.StatusCode, rr.BodyBase64)
}
