package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	ErrResourceNotFound = "resource not found: %s"
)

func FetchUrl(url *url.URL) ([]byte, error) {

	res, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(ErrResourceNotFound, url.String())
	}

	// Read the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
