package internal

import (
	"net/http"
	"time"
)

// HTTPClient is an interface over http.Client to make mocking easier.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// httpClientWithHeader wraps an HTTP client to add custom headers like "secret".
type httpClientWithHeader struct {
	client HTTPClient
	secret string
}

// Do method wraps the original client's Do method and injects the "secret" header.
func (h *httpClientWithHeader) Do(req *http.Request) (*http.Response, error) {
	// Add the secret header before sending the request
	req.Header.Set("secret", h.secret)
	return h.client.Do(req)
}

// DefaultHTTPClient returns a default HTTP client with a 10-second timeout.
func DefaultHTTPClient() HTTPClient {
	return HTTPClientWithTimeout(10 * time.Second)
}

// HTTPClientWithTimeout returns an HTTP client with a custom timeout.
func HTTPClientWithTimeout(timeout time.Duration) HTTPClient {
	httpClient := http.DefaultClient
	httpClient.Timeout = timeout
	return httpClient
}

// HTTPClientWithSecret wraps an HTTP client and adds the "secret" header.
func HTTPClientWithSecret(client HTTPClient, secret string) HTTPClient {
	return &httpClientWithHeader{
		client: client,
		secret: secret,
	}
}
