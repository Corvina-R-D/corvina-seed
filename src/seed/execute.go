package seed

import (
	"context"
	"corvina/corvina-seed/src/seed/api"
	"corvina/corvina-seed/src/seed/dto"
	"math/rand"

	"github.com/lucasepe/codename"
	"github.com/rs/zerolog/log"
)

func Execute(ctx context.Context, input *dto.ExecuteInDTO) error {

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

	err = createModels(ctx, input, organization, rng)
	if err != nil {
		return err
	}

	return nil
}

func createDeviceGroups(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO, rng *rand.Rand) error {
	for i := int64(0); i < input.DeviceGroupCount; i++ {
		err := api.CreateDeviceGroup(ctx, organization.Id, api.CreateDeviceGroupInDTO{
			Name: codename.Generate(rng, 4),
		})
		if err != nil {
			return err
		}
	}

	log.Info().Int64("device group count", input.DeviceGroupCount).Msg("Device groups created")

	return nil
}

func createModels(ctx context.Context, input *dto.ExecuteInDTO, organization *dto.OrganizationOutDTO, rng *rand.Rand) error {
	for i := int64(0); i < input.ModelCount; i++ {
		name := codename.Generate(rng, 4) + ":1.0.0"
		output, err := api.CreateModel(ctx, organization.ResourceID, api.CreateModelInDTO{
			Name: name,
			Data: api.ModelDataDTO{
				Type:       "object",
				InstanceOf: name,
				Properties: map[string]api.CreateModelInDataPropertiesDTO{
					"temperature": {
						Type: "double",
					},
					"humidity": {
						Type: "boolean",
					},
					"description": {
						Type: "string",
					},
				},
				Label:       codename.Generate(rng, 4),
				Unit:        "°C",
				Description: codename.Generate(rng, 4),
				Tags:        []string{codename.Generate(rng, 4), codename.Generate(rng, 4)},
			},
		})
		if err != nil {
			return err
		}

		log.Debug().Str("model.id", output.Id).Msg("Model created")
	}

	log.Info().Int64("model count", input.ModelCount).Msg("Models created")

	return nil
}
