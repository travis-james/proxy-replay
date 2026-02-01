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
	headers := func() map[string][]string {
		retval := map[string][]string{}
		for key, val := range resp.Header {
			if _, ok := HeadersToRecord[key]; !ok {
				continue
			}
			retval[key] = val
		}
		return retval
	}()

	rr := RecordedResponse{
		StatusCode: resp.StatusCode,
		BodyBase64: encodedBody,
		Headers:    headers,
	}

	got, err := rr.ToJSON()
	if err != nil {
		return err
	}
	fmt.Println(string(got))
	return nil
}
