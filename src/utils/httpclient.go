package utils

import (
	"net/http"
	"time"
)

var HttpClient http.Client = http.Client{
	Timeout: 30 * time.Second,
}

var HttpClientNoFollow http.Client = http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}
