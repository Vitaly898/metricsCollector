package bodyclose

import "net/http"

func bad() error {
	response, err := http.Get("http://example.com") // want "response body must be closed"
	if err != nil {
		return err
	}
	_ = response.StatusCode
	return nil
}
