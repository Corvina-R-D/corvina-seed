package cmd

import (
	"context"
	"corvina/corvina-seed/src/seed/api"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/utils"
	"os"

	"github.com/rs/zerolog/log"
)

func DeviceAuthz(ctx context.Context) error {

	name := utils.RandomName()
	folderName := ctx.Value(utils.DomainKey).(string) + "." + name
	if err := os.Mkdir(folderName, os.ModePerm); err != nil {
		return err
	}

	organization, err := api.GetOrganizationMine(ctx)
	if err != nil {
		return err
	}
	log.Info().Interface("organization", organization).Msg("Organization retrieved")

	err = api.CreateDevice(ctx, organization.ResourceID, name)
	if err != nil {
		return err
	}
	log.Info().Str("device name", name).Msg("Device created")

	err = CreateServiceAccountWithDeviceAssociated(ctx, organization, name)
	if err != nil {
		return err
	}

	// TODO: put certificate in the folder

	return nil
}

func CreateServiceAccountWithDeviceAssociated(ctx context.Context, organization *dto.OrganizationOutDTO, deviceName string) error {
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
