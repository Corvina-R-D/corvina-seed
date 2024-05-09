package seed

import (
	"context"
	"corvina/corvina-seed/src/seed/api"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/utils"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

func Execute(ctx context.Context, input *dto.ExecuteInDTO) error {

	organization, err := api.GetOrganizationMine(ctx)
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

	err = createServiceAccounts(ctx, input, organization)
	if err != nil {
		return err
	}

	err = createOrganizations(ctx, input, organization.ID)
	if err != nil {
		return err
	}

	return nil
}

func createServiceAccounts(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO) error {
	if input.ServiceAccountCount == 0 {
		return nil
	}

	for i := int64(0); i < input.ServiceAccountCount; i++ {
		_, err := api.CreateServiceAccount(ctx, organization.ID, utils.RandomName())
		if err != nil {
			return err
		}
	}

	utils.PrintlnGreen(fmt.Sprintf("Service accounts created: %d", input.ServiceAccountCount))

	return nil
}

func createDevices(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO) error {
	if input.DeviceCount == 0 {
		return nil
	}

	for i := int64(0); i < input.DeviceCount; i++ {
		_, err := api.CreateDevice(ctx, organization.ResourceID, utils.RandomName())
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
		err := api.CreateDeviceGroup(ctx, organization.ID, api.CreateDeviceGroupInDTO{
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

func calculateNumberOfOrganizations(input *dto.ExecuteInDTO) int64 {
	sum := int64(0)

	for i := int64(0); i <= input.OrganizationTreeDepth; i++ {
		sum += utils.PowInt64(input.OrganizationCount, i)
	}

	return sum
}

func createOrganizations(ctx context.Context, input *dto.ExecuteInDTO, organizationId int64) error {
	if input.OrganizationCount == 0 {
		return nil
	}

	totalOrganizations := calculateNumberOfOrganizations(input)
	utils.PrintlnBlue(fmt.Sprintf("Creating organizations: %d with depth %d = %d", input.OrganizationCount, input.OrganizationTreeDepth, totalOrganizations))

	err := createSubOrgMultipleWorkers(ctx, input, organizationId)
	if err != nil {
		return err
	}

	utils.PrintlnGreen(fmt.Sprintf("Organizations created: %d", totalOrganizations))

	return nil
}

type WorkerDTO struct {
	depth  int64
	parent int64
}

func createSubOrgMultipleWorkers(ctx context.Context, input *dto.ExecuteInDTO, organizationId int64) error {
	depth := input.OrganizationTreeDepth
	leafs := input.OrganizationCount
	numOfDepthWorkers := 20
	var c = make(chan WorkerDTO, utils.PowInt64(input.OrganizationCount, input.OrganizationTreeDepth))
	wg := sync.WaitGroup{}
	defer func() {
		wg.Wait()
		close(c)
	}()

	wg.Add(1)
	c <- WorkerDTO{depth: depth, parent: organizationId}

	for i := 0; i < numOfDepthWorkers; i++ {
		go func(workId int) {
			for dto := range c {
				if dto.depth == 0 {
					wg.Done()
					continue
				}

				for i := int64(0); i < leafs; i++ {
					subOrganization, err := createSubOrg(ctx, dto.parent)
					if err != nil {
						log.Error().Err(err).Int64("organziationId", organizationId).Msg("Error creating sub organization")
						continue
					}

					log.Debug().Msgf("Worker %d, parent %d, depth %d\n", workId, dto.parent, depth-dto.depth+1)
					if dto.depth > 0 {
						wg.Add(1)
						c <- WorkerDTO{depth: dto.depth - 1, parent: subOrganization.ID}
					}
				}
				wg.Done()
			}
		}(i)
	}

	return nil
}

func createSubOrg(ctx context.Context, organizationId int64) (*api.CreateOrganizationOutDTO, error) {
	createSubOrgInput := api.CreateOrganizationInDTO{
		Name: utils.RandomName(),
	}
	subOrganization, err := api.CreateSubOrganization(ctx, organizationId, createSubOrgInput)
	if err != nil {
		return nil, err
	}
	err = api.SetAllLimitToUnlimited(ctx, subOrganization.ResourceID)
	if err != nil {
		return nil, err
	}
	return subOrganization, err
}
