package cmd

import (
	"context"
	"corvina/corvina-seed/src/seed"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/utils"
	"errors"

	"github.com/rs/zerolog/log"
)

func Run(ctx context.Context) error {

	origin := ctx.Value(utils.OriginKey).(string)
	log.Debug().Str("origin", origin).Msg("")

	deviceCount, err := takeCountFromCtx(ctx, utils.DeviceCount)
	if err != nil {
		return err
	}
	log.Debug().Int64("device count", deviceCount).Msg("")

	deviceGroupCount, err := takeCountFromCtx(ctx, utils.DeviceGroupCount)
	if err != nil {
		return err
	}
	log.Debug().Int64("device group count", deviceGroupCount).Msg("")

	modelCount, err := takeCountFromCtx(ctx, utils.ModelCount)
	if err != nil {
		return err
	}
	log.Debug().Int64("model count", modelCount).Msg("")

	serviceAccountCount, err := takeCountFromCtx(ctx, utils.ServiceAccountCount)
	if err != nil {
		return err
	}
	log.Debug().Int64("service account count", serviceAccountCount).Msg("")

	organizationCount, err := takeCountFromCtx(ctx, utils.OrganizationCount)
	if err != nil {
		return err
	}
	log.Debug().Int64("organization count", organizationCount).Msg("")
	organizationTreeDepth, err := takeCountFromCtx(ctx, utils.OrganizationTreeDepth)
	if err != nil {
		return err
	}
	log.Debug().Int64("organization tree depth", organizationTreeDepth).Msg("")

	executeInput := dto.ExecuteInDTO{
		DeviceCount:           deviceCount,
		DeviceGroupCount:      deviceGroupCount,
		ModelCount:            modelCount,
		ServiceAccountCount:   serviceAccountCount,
		OrganizationCount:     organizationCount,
		OrganizationTreeDepth: organizationTreeDepth,
	}

	if !atLeastOneCountIsProvided(executeInput) {
		utils.PrintlnYellow("No count provided, nothing to do, try --help to understand how to use it")
		return nil
	}

	err = seed.Execute(ctx, &executeInput)
	if err != nil {
		return err
	}

	return nil
}

func atLeastOneCountIsProvided(counters dto.ExecuteInDTO) bool {
	return counters.DeviceCount > 0 || counters.DeviceGroupCount > 0 || counters.ModelCount > 0 || counters.OrganizationCount > 0 || counters.ServiceAccountCount > 0
}

func takeCountFromCtx(ctx context.Context, ctxKey utils.CtxKey) (int64, error) {
	val := ctx.Value(ctxKey)

	if val == nil {
		log.Trace().Interface("key", ctxKey).Msg("no count value provided")
		return 0, nil
	}

	count := val.(int64)

	if count < 0 {
		return 0, errors.New("invalid count value")
	}

	return count, nil
}
