package cloud

import "net/http"

// Extract the client creation for UT purposes

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
