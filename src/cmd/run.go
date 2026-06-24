package cmd

import (
	"context"
	"corvina/corvina-seed/src/seed"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/utils"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
)

func Run(ctx context.Context) error {

	utils.PrintlnBlue("Started at: " + time.Now().Format("2006-01-02 15:04:05"))

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

	userCount, err := takeCountFromCtx(ctx, utils.UserCount)
	if err != nil {
		return err
	}
	log.Debug().Int64("user count", userCount).Msg("")

	// SeedOrgName è una stringa, non un contatore: la leggiamo direttamente dal context.
	seedOrgName, _ := ctx.Value(utils.SeedOrgName).(string)
	log.Debug().Str("seed org name", seedOrgName).Msg("")

	orgWorkers, err := takeCountFromCtx(ctx, utils.OrgWorkers)
	if err != nil {
		return err
	}
	// Default a 1 worker: serializza completamente le chiamate così il rate
	// limiter del server non viene mai saturato. Aumentare solo su ambienti
	// con rate limit più alto e confermato empiricamente.
	if orgWorkers == 0 {
		orgWorkers = 1
	}
	log.Debug().Int64("org workers", orgWorkers).Msg("")

	requestDelayMs, err := takeCountFromCtx(ctx, utils.RequestDelayMs)
	if err != nil {
		return err
	}
	// Default a 200ms: con 3 worker equivale a ~15 req/s, abbondantemente
	// sotto il rate limit tipico e sufficiente per alberi di centinaia di org.
	if requestDelayMs == 0 {
		requestDelayMs = 200
	}
	log.Debug().Int64("request delay ms", requestDelayMs).Msg("")

	executeInput := dto.ExecuteInDTO{
		DeviceCount:           deviceCount,
		DeviceGroupCount:      deviceGroupCount,
		ModelCount:            modelCount,
		ServiceAccountCount:   serviceAccountCount,
		OrganizationCount:     organizationCount,
		OrganizationTreeDepth: organizationTreeDepth,
		UserCount:             userCount,
		SeedOrgName:           seedOrgName,
		OrgWorkers:            orgWorkers,
		RequestDelayMs:        requestDelayMs,
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
	// SeedOrgName da solo è sufficiente: crea la org contenitore e stampa il resource_id.
	return counters.DeviceCount > 0 || counters.DeviceGroupCount > 0 || counters.ModelCount > 0 || counters.OrganizationCount > 0 || counters.ServiceAccountCount > 0 || counters.UserCount > 0 || counters.SeedOrgName != ""
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
