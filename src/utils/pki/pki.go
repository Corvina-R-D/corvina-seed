package pki

import (
	"fmt"
	"os"
	"path/filepath"

	"corvina/corvina-seed/src/utils"

	"github.com/rs/zerolog/log"
)

const PrivateKeyRelativePath = "own/private/cert.key"
const CertificateRelativePath = "own/certs/cert.crt"

var PkiRoot = "./." + utils.RandomName() + "-pki"

func SetupPKIFolder() {
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
		err := os.MkdirAll(PkiPath(dir), 0755)
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("Error creating directory %s:\n", dir))
			return
		}
	}

}

func CleanPKIFolder() error {
	return os.RemoveAll(PkiRoot)
}

func PkiPath(relativePath string) string {
	return filepath.Join(PkiRoot, relativePath)
}

func CertificatePath() string {
	return PkiPath(CertificateRelativePath)
}

func PrivateKeyPath() string {
	return PkiPath(PrivateKeyRelativePath)
}
