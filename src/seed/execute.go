package seed

import (
	"context"
	"corvina/corvina-seed/src/seed/api"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/utils"
	"fmt"

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

	err = createOrganizations(ctx, input, organization.Id)
	if err != nil {
		return err
	}

	return nil
}

func createDevices(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO) error {
	if input.DeviceCount == 0 {
		return nil
	}

	for i := int64(0); i < input.DeviceCount; i++ {
		err := api.CreateDevice(ctx, organization.ResourceID, utils.RandomName())
		if err != nil {
			return err
		}
	}

	utils.PrintlnGreen(fmt.Sprintf("Devices created: %d", input.DeviceCount))

	return nil
}

func createDeviceGroups(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO) error {
	if input.DeviceGroupCount == 0 {
		return nil
	}

	for i := int64(0); i < input.DeviceGroupCount; i++ {
		err := api.CreateDeviceGroup(ctx, organization.Id, api.CreateDeviceGroupInDTO{
			Name: utils.RandomName(),
		})
		if err != nil {
			return err
		}
	}

	utils.PrintlnGreen(fmt.Sprintf("Device groups created: %d", input.DeviceGroupCount))

	return nil
}

func createModels(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO) error {
	if input.ModelCount == 0 {
		return nil
	}

	for i := int64(0); i < input.ModelCount; i++ {
		output, err := api.CreateRandomModel(ctx, organization.ResourceID)
		if err != nil {
			return err
		}

		log.Debug().Str("model.id", output.Id).Msg("Model created")
	}

	utils.PrintlnGreen(fmt.Sprintf("Models created: %d", input.ModelCount))

	return nil
}

func createOrganizations(ctx context.Context, input *dto.ExecuteInDTO, organizationId int64) error {
	if input.OrganizationCount == 0 {
		return nil
	}

	utils.PrintlnBlue(fmt.Sprintf("Creating organizations: %d with depth %d = %d", input.OrganizationCount, input.OrganizationTreeDepth, utils.PowInt64(input.OrganizationCount, input.OrganizationTreeDepth)))

	err := createSubOrgRecursively(ctx, input, organizationId)
	if err != nil {
		return err
	}

	utils.PrintlnGreen(fmt.Sprintf("Organizations created: %d with depth %d", input.OrganizationCount, input.OrganizationTreeDepth))

	return nil
}

func createSubOrgRecursively(ctx context.Context, input *dto.ExecuteInDTO, organizationId int64) error {
	for i := int64(0); i < input.OrganizationCount; i++ {
		createSubOrgInput := api.CreateOrganizationInDTO{
			Name: utils.RandomName(),
		}
		subOrganization, err := api.CreateSubOrganization(ctx, organizationId, createSubOrgInput)
		if err != nil {
			return err
		}
		err = api.SetAllLimitToUnlimited(ctx, subOrganization.ResourceID)
		if err != nil {
			return err
		}

		log.Debug().Interface("organization", subOrganization.ResourceID).Msg("Organization created")

		if input.OrganizationTreeDepth > 1 {
			executeInput := &dto.ExecuteInDTO{
				OrganizationCount:     input.OrganizationCount,
				OrganizationTreeDepth: input.OrganizationTreeDepth - 1,
			}
			err = createSubOrgRecursively(ctx, executeInput, subOrganization.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
