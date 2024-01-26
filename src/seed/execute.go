package seed

import (
	"context"
	"corvina/corvina-seed/src/seed/api"
	"corvina/corvina-seed/src/seed/dto"
	"math/rand"

	"github.com/lucasepe/codename"
	"github.com/rs/zerolog/log"
)

func Execute(ctx context.Context, input dto.ExecuteInDTO) error {

	organization, err := api.GetOrganizationMine(ctx)
	if err != nil {
		return err
	}

	rng, err := codename.DefaultRNG()
	if err != nil {
		return err
	}
	log.Info().Interface("organization", organization).Msg("Organization retrieved")

	err = createDeviceGroups(ctx, input, organization, rng)
	if err != nil {
		return err
	}

	return nil
}

func createDeviceGroups(ctx context.Context, input dto.ExecuteInDTO, organization dto.OrganizationOutDTO, rng *rand.Rand) error {
	for i := int64(0); i < input.DeviceGroupCount; i++ {
		err := api.CreateDeviceGroup(ctx, organization.ID, api.CreateDeviceGroupInDTO{
			Name: codename.Generate(rng, 4),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
