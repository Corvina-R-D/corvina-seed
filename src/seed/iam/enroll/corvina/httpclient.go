package corvina

import (
	"corvina/corvina-seed/src/seed/iam/certificates"
	"crypto/tls"
	"net/http"
)

var Client *http.Client

func init() {

	// Create a new TLS configuration with the certificate pool
	tlsConfig := &tls.Config{
		RootCAs:            certificates.OtaCaCertPool,
		InsecureSkipVerify: true,
	}

	// Create a new transport with the TLS configuration
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Create a new HTTP client with the transport
	Client = &http.Client{
		Transport: transport,
	}
}
