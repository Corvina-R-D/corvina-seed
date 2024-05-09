package cmd

import (
	"context"
	"corvina/corvina-seed/src/seed/api"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/seed/iam/certificates"
	"corvina/corvina-seed/src/seed/iam/enroll"
	"corvina/corvina-seed/src/utils"
	"corvina/corvina-seed/src/utils/pki"
	"fmt"
	"os"
	"path"

	"github.com/rs/zerolog/log"
)

func DeviceAuthz(ctx context.Context) error {

	certificates.InitializeCertificate()

	name := utils.RandomName()

	folderName, err := createDeviceFolder(ctx, name)
	if err != nil {
		return err
	}
	log.Info().Str("folder name", *folderName).Msg("Device folder created")

	organization, err := api.GetOrganizationMine(ctx)
	if err != nil {
		return err
	}
	log.Info().Interface("organization", organization).Msg("Organization retrieved")

	activationKey, err := api.CreateDevice(ctx, organization.ResourceID, name)
	if err != nil {
		return err
	}
	log.Info().Str("device name", name).Msg("Device created")

	err = createServiceAccountWithDeviceAssociated(ctx, organization, name)
	if err != nil {
		return err
	}

	_, _, err = enroll.CorvinaEnroll("https://pairing.corvina.mk/api/v1/", *activationKey)
	if err != nil {
		return err
	}

	err = saveCertificate(*folderName)
	if err != nil {
		return err
	}
	log.Info().Str("folder name", *folderName).Msg("Certificate saved")

	err = savePrivateKey(*folderName)
	if err != nil {
		return err
	}
	log.Info().Str("folder name", *folderName).Msg("Private key saved")

	err = pki.CleanPKIFolder()
	if err != nil {
		return err
	}
	log.Info().Msg("PKI folder cleaned")

	utils.PrintlnGreen(fmt.Sprintf("Folder %s created!", *folderName))
	utils.PrintlnBlue(fmt.Sprintf(`
cd %s
curl https://%s/svc/core/api/v1/users/mine --cert cert.crt --key cert.key
	`, *folderName, deviceDomain(ctx)))

	return nil
}

func deviceDomain(ctx context.Context) string {
	rootDomain := ctx.Value(utils.DomainKey).(string)
	if rootDomain == "corvina.mk" {
		rootDomain = "corvina.mk" + ":8443"
	}

	return "device." + rootDomain
}

func saveCertificate(folderName string) error {
	return utils.CopyFile(pki.CertificatePath(), path.Join(folderName, "cert.crt"))
}

func savePrivateKey(folderName string) error {
	return utils.CopyFile(pki.PrivateKeyPath(), path.Join(folderName, "cert.key"))
}

func createDeviceFolder(ctx context.Context, name string) (*string, error) {
	folderName := ctx.Value(utils.DomainKey).(string) + "." + name
	if err := os.Mkdir(folderName, os.ModePerm); err != nil {
		return nil, err
	}
	return &folderName, nil
}

func createServiceAccountWithDeviceAssociated(ctx context.Context, organization *dto.OrganizationOutDTO, deviceName string) error {
	adminRole, err := api.GetFirstAdminApplicationRole(ctx, organization.ID)
	if err != nil {
		return err
	}
	log.Debug().Int64("admin role", adminRole.ID).Msg("Admin role retrieved")
	adminDeviceRole, err := api.GetFirstAdminDeviceRole(ctx, organization.ID)
	if err != nil {
		return err
	}
	log.Debug().Int64("admin device role", adminDeviceRole.ID).Msg("Admin device role retrieved")
	user, err := api.CreateServiceAccount(ctx, organization.ID, deviceName)
	if err != nil {
		return err
	}
	log.Info().Interface("user", user).Msg("Service account created with the same name as the device")
	roles := []int64{adminRole.ID, adminDeviceRole.ID}
	err = api.AssignRolesToUser(ctx, organization.ID, int64(user.ID), roles)
	if err != nil {
		return err
	}
	log.Info().Interface("roles", roles).Msg("Roles assigned to user")
	securityPolicy, err := api.GetSecurityPolicy(ctx, organization.ID, deviceName)
	if err != nil {
		return err
	}
	err = api.AssignSecurityPolicyToUser(ctx, organization.ID, securityPolicy.ID, user.ID)
	if err != nil {
		return err
	}
	log.Info().Interface("security policy", securityPolicy).Msg("Security policy assigned to user")

	return nil
}
