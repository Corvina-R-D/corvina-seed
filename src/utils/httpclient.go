package utils

import (
	"crypto/tls"
	"net/http"
	"time"
)

var defaultTransport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var HttpClient http.Client = http.Client{
	Timeout:   30 * time.Second,
	Transport: defaultTransport,
}

var HttpClientNoFollow http.Client = http.Client{
	Timeout:   30 * time.Second,
	Transport: defaultTransport,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}
