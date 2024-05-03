package cmd

import (
	"context"
	"corvina/corvina-seed/src/seed/api"
	"corvina/corvina-seed/src/utils"
	"os"

	"github.com/rs/zerolog/log"
)

func DeviceAuthz(ctx context.Context) error {

	deviceName := utils.RandomName()
	folderName := ctx.Value(utils.DomainKey).(string) + "." + deviceName
	if err := os.Mkdir(folderName, os.ModePerm); err != nil {
		return err
	}

	organization, err := api.GetOrganizationMine(ctx)
	if err != nil {
		return err
	}
	log.Info().Interface("organization", organization).Msg("Organization retrieved")

	err = api.CreateDevice(ctx, organization.ResourceID, deviceName)
	if err != nil {
		return err
	}
	log.Info().Str("device name", deviceName).Msg("Device created")

	// TODO: create a service account with this device associated
	adminRole, err := api.GetFirstAdminApplicationRole(ctx, organization.Id)
	if err != nil {
		return err
	}
	log.Debug().Int64("admin role", adminRole.ID).Msg("Admin role retrieved")
	adminDeviceRole, err := api.GetFirstAdminDeviceRole(ctx, organization.Id)
	if err != nil {
		return err
	}
	log.Debug().Int64("admin device role", adminDeviceRole.ID).Msg("Admin device role retrieved")

	// TODO: put certificate in the folder

	return nil
}
