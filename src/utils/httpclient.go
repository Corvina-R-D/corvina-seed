package utils

import (
	"net/http"
	"time"
)

var HttpClient http.Client = http.Client{
	Timeout: 30 * time.Second,
}
