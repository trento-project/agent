package cloud

import "net/http"

// Extract the client creation for UT purposes
//go:generate mockery --name=HTTPClient
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var client HTTPClient = &http.Client{Transport: &http.Transport{Proxy: nil}} //nolint
