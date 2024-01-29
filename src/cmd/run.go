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

	executeInput := dto.ExecuteInDTO{
		DeviceCount:      deviceCount,
		DeviceGroupCount: deviceGroupCount,
		ModelCount:       modelCount,
	}
	err = seed.Execute(ctx, &executeInput)
	if err != nil {
		return err
	}

	return nil
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
