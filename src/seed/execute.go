package seed

import (
	"context"
	"corvina/corvina-seed/src/seed/api"
	"corvina/corvina-seed/src/seed/dto"

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

	for i := int64(0); i < input.DeviceGroupCount; i++ {
		err = api.CreateDeviceGroup(ctx, organization.ID, api.CreateDeviceGroupInDTO{
			Name: codename.Generate(rng, 4),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
