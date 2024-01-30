package seed

import (
	"context"
	"corvina/corvina-seed/src/seed/api"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/utils"

	"github.com/rs/zerolog/log"
)

func Execute(ctx context.Context, input *dto.ExecuteInDTO) error {

	organization, err := api.GetOrganizationMine(ctx)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	log.Info().Interface("organization", organization).Msg("Organization retrieved")

	err = createDeviceGroups(ctx, input, organization)
	if err != nil {
		return err
	}

	err = createModels(ctx, input, organization)
	if err != nil {
		return err
	}

	err = createDevices(ctx, input, organization)
	if err != nil {
		return err
	}

	return nil
}

func createDevices(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO) error {
	for i := int64(0); i < input.DeviceCount; i++ {
		err := api.CreateDevice(ctx, organization.ResourceID, utils.RandomName())
		if err != nil {
			return err
		}
	}

	log.Info().Int64("device count", input.DeviceCount).Msg("Devices created")

	return nil
}

func createDeviceGroups(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO) error {
	for i := int64(0); i < input.DeviceGroupCount; i++ {
		err := api.CreateDeviceGroup(ctx, organization.Id, api.CreateDeviceGroupInDTO{
			Name: utils.RandomName(),
		})
		if err != nil {
			return err
		}
	}

	log.Info().Int64("device group count", input.DeviceGroupCount).Msg("Device groups created")

	return nil
}

func createModels(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO) error {
	for i := int64(0); i < input.ModelCount; i++ {
		output, err := api.CreateRandomModel(ctx, organization.ResourceID)
		if err != nil {
			return err
		}

		log.Debug().Str("model.id", output.Id).Msg("Model created")
	}

	log.Info().Int64("model count", input.ModelCount).Msg("Models created")

	return nil
}
