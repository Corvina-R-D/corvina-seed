package pki

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

const OtaPrivateKeyRelativePath = "own/private/cert.key"
const OtaCertificateRelativePath = "own/certs/cert.crt"
const OTAPkiRoot = "./ota-pki"

func SetupOtaPKIFolder() {
	// Define the directories to be created
	dirs := []string{
		"issuers/certs",
		"issuers/crl",
		"own/certs",
		"own/private",
		"rejected",
		"trusted/certs",
		"trusted/crl",
	}

	// Create each directory
	for _, dir := range dirs {
		err := os.MkdirAll(OtaPkiPath(dir), 0755)
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("Error creating directory %s:\n", dir))
			return
		}
	}

}

func OtaPkiPath(relativePath string) string {
	return filepath.Join(OTAPkiRoot, relativePath)
}
