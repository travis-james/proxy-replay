package recorder

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

// build and send http request
// capture response
// produce recorded object.

func Record(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	encodedBody := base64.StdEncoding.EncodeToString(body)

	rr := RecordedResponse{
		StatusCode: resp.StatusCode,
		BodyBase64: encodedBody,
		Headers:    resp.Header,
	}

	fmt.Println(rr.Stringify(true))
	return nil
}
