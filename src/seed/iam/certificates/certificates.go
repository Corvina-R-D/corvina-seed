package certificates

import (
	"corvina/corvina-seed/src/utils/pki"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
)

var OtaCaCertPool *x509.CertPool
var OtaCertificates []tls.Certificate

func LoadOtaCertificates() (err error) {
	cert, err := tls.LoadX509KeyPair(pki.PkiPath(pki.CertificateRelativePath), pki.PkiPath(pki.PrivateKeyRelativePath))
	if err != nil {
		log.Info().Err(err).Msg("Error loading certificate or private key, please enroll again")
		return err
	}

	// Parse the certificate
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return err
	}

	// Check if the certificate is going to expire
	if time.Now().After(x509Cert.NotAfter) {
		return errors.New("certificate has expired, please enroll again")
	} else if time.Now().Add(2 * 24 * time.Hour).After(x509Cert.NotAfter) {
		log.Warn().Msg("Certificate will expire in less than 2 days")
		return nil
	} else {
		log.Debug().Time("NotAfter", x509Cert.NotAfter).Msg("Certificate will expire")
	}

	OtaCertificates = []tls.Certificate{cert}
	return nil
}

func InitializeCertificate() {
	pki.SetupPKIFolder()

	// initialize vars, if files are available
	// Read the CA certificate
	caCert := []byte(`-----BEGIN CERTIFICATE-----
MIICWTCCAf+gAwIBAgIUAkkMEwP0AejpBDLeXUiBRJSDv7UwCgYIKoZIzj0EAwIw
eTELMAkGA1UEBhMCSVQxDjAMBgNVBAgMBUl0YWx5MSgwJgYDVQQKDB9FeG9yIERl
dmljZXMgRGlnaXRhbCBJZGVudGl0aWVzMTAwLgYDVQQDDCdFeG9yIERldmljZXMg
RGlnaXRhbCBJZGVudGl0aWVzIFJvb3QgQ0EwIBcNMjAxMjEwMTAwOTQ5WhgPMjA2
MjAxMDQxMDA5NDlaMHkxCzAJBgNVBAYTAklUMQ4wDAYDVQQIDAVJdGFseTEoMCYG
A1UECgwfRXhvciBEZXZpY2VzIERpZ2l0YWwgSWRlbnRpdGllczEwMC4GA1UEAwwn
RXhvciBEZXZpY2VzIERpZ2l0YWwgSWRlbnRpdGllcyBSb290IENBMFkwEwYHKoZI
zj0CAQYIKoZIzj0DAQcDQgAEQGKIj1KpHpRk5ZOYvf9g33ENs2gOBu3RsCneaYKQ
Jhhl8wzVnt8vA4wzgv7B9Jui5+efYIk9N19jZ9H8JAjDZKNjMGEwHQYDVR0OBBYE
FO3l09dQYmSZ5+VuR8IDyNDSrP8cMB8GA1UdIwQYMBaAFO3l09dQYmSZ5+VuR8ID
yNDSrP8cMA8GA1UdEwEB/wQFMAMBAf8wDgYDVR0PAQH/BAQDAgGGMAoGCCqGSM49
BAMCA0gAMEUCIEBfvBPKnQSGQhk/JLvtdsC9AUhzmpnmXKqztImkkkfJAiEAqEOc
fLibdXgfUjlbFwApfXoXZsYZMwyFq/HjIKS1pyA=
-----END CERTIFICATE-----`)

	var err error
	// Create a new certificate pool
	OtaCaCertPool, err = x509.SystemCertPool()
	if err != nil {
		log.Error().Err(err).Msg("Could not load system certificates")
		OtaCaCertPool = x509.NewCertPool()
	}

	// Append the certificate to the pool
	if ok := OtaCaCertPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatal().Msg("Adding Corvina CA certificate to pool failed")
	}

	// Load identity if any
	LoadOtaCertificates()

}
