package cache

import (
	"fmt"

	"io"
	"net/http"
	"net/url"
)

var _ Fetcher = (*httpFetcher)(nil)

type httpFetcher struct {
	baseURL string
}

// httpFetcher responsible for querying the value of key from the group cache of the specified node through http request
func (h *httpFetcher) Fetch(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))

	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %v", err)
	}

	return b, nil
}
