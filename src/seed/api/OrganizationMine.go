package api

import (
	"context"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

func GetOrganizationMine(ctx context.Context) (dto.OrganizationOutDTO, error) {
	origin := ctx.Value(utils.OriginKey).(string)
	apiKey := ctx.Value(utils.ApiKey).(string)

	url := origin + "/svc/core/api/v1/organizations/mine"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return dto.OrganizationOutDTO{}, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("X-Api-Key", apiKey)

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return dto.OrganizationOutDTO{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return dto.OrganizationOutDTO{}, err
	}

	if resp.StatusCode != 200 {
		return dto.OrganizationOutDTO{}, errors.New("Error retrieving organization. Status:" + resp.Status + ". Body: " + string(body))
	}

	log.Trace().Str("body", string(body)).Msg("/svc/core/api/v1/organizations/mine response")

	var organizations []dto.OrganizationOutDTO
	err = json.Unmarshal(body, &organizations)
	if err != nil {
		return dto.OrganizationOutDTO{}, err
	}

	return organizations[0], nil
}
