package seed

import (
	"context"
	"corvina/corvina-seed/src/seed/api"
	"corvina/corvina-seed/src/seed/dto"

	"github.com/rs/zerolog/log"
)

func Execute(ctx context.Context, input dto.ExecuteInDTO) error {

	organization, err := api.GetOrganizationMine(ctx)
	if err != nil {
		return err
	}

	log.Info().Interface("organization", organization).Msg("Organization retrieved")

	return nil
}
