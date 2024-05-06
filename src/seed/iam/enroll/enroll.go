package enroll

import (
	"corvina/corvina-seed/src/seed/iam/enroll/corvina"
	"corvina/corvina-seed/src/utils/pki"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type DeviceIdentifiers struct {
	DeviceId       string
	InstanceId     string
	OrganizationId string
}

func createCSR(logicalId string) (string, error) {
	// check if private key is present in pki folder own/private

	// if private key is not present in pki folder own/private, generate a new one

	var privateKey *rsa.PrivateKey
	privateKeyPath := filepath.Join(pki.OTAPkiRoot, pki.OtaPrivateKeyRelativePath)

	_, err := os.Stat(privateKeyPath)
	if os.IsNotExist(err) {
		_privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

		if err != nil {
			return "", err
		}
		privateKey = _privateKey

		// save the key to file

		// Encode the private key into a PEM block
		privateKeyBlock := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}

		// Open the private key file
		privateKeyFile, err := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return "", err
		}
		defer privateKeyFile.Close()

		// Write the PEM block to the file
		err = pem.Encode(privateKeyFile, privateKeyBlock)
		if err != nil {
			return "", err
		}
	} else {
		privateKeyBytes, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return "", err
		}

		block, _ := pem.Decode(privateKeyBytes)
		if block == nil {
			return "", errors.New("failed to decode PEM block containing private key")
		}

		_privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", err
		}
		privateKey = _privateKey
	}

	template := x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"System"},
			CommonName:   logicalId,
		},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return "", err
	}

	csr := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	return string(csr), nil
}

func CorvinaEnroll(pairingUri string, activationKey string) (keyPair *tls.Certificate, device DeviceIdentifiers, err error) {

	// Initialize the LicensesService
	licensesService := corvina.NewLicensesService(pairingUri, activationKey)

	// Initialize the LicensesService
	licenseData, err := licensesService.Init()
	device = DeviceIdentifiers{
		DeviceId:       licenseData.LogicalId,
		InstanceId:     licenseData.InstanceId,
		OrganizationId: licenseData.OrganizationId,
	}
	if err != nil {
		log.Error().Err(err).Msg("Error initializing LicensesService")
		return nil, device, err
	}

	// Create a CSR
	csr, err := createCSR(licenseData.LogicalId)
	if err != nil {
		log.Error().Err(err).Msg("Error creating CSR")
		return nil, device, err
	}

	// Do the pairing
	crt, err := licensesService.DoPairing(csr)
	if err != nil {
		log.Error().Err(err).Msg("Error doing pairing")
		return nil, device, err
	}

	log.Info().Str("crt", crt.Data.ClientCrt).Msg("CSR created")

	// Verify the certificate
	verified, err := licensesService.Verify(crt.Data.ClientCrt)
	if err != nil {
		log.Error().Err(err).Msg("Error verifying certificate")
		return nil, device, err
	}

	if verified {
		// save certificate to pki folder own/certs
		certPath := pki.OtaPkiPath(pki.OtaCertificateRelativePath)
		certFile, err := os.Create(certPath)
		if err != nil {
			log.Error().Err(err).Msg("Error creating certificate file")
			return nil, device, err
		}
		defer certFile.Close()

		_, err = certFile.WriteString(crt.Data.ClientCrt)
		if err != nil {
			log.Error().Err(err).Msg("Error writing certificate to file")
			return nil, device, err
		}
	}

	log.Info().Bool("Certificate verified", verified).Msg("Certificate verified")

	cert, err := tls.LoadX509KeyPair(pki.OtaPkiPath(pki.OtaCertificateRelativePath), pki.OtaPkiPath(pki.OtaPrivateKeyRelativePath))
	if err != nil {
		log.Error().Err(err).Msg("Error loading certificate")
		return nil, device, err
	}

	return &cert, device, nil
}
