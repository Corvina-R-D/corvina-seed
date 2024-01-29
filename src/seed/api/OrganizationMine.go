package api

import (
	"context"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

func GetOrganizationMine(ctx context.Context) (*dto.OrganizationOutDTO, error) {
	origin := ctx.Value(utils.OriginKey).(string)

	url := origin + "/svc/core/api/v1/organizations/mine"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	token, err := keycloak.AdminToken(ctx)
	if err != nil {
		return nil, err
	}
	req.Header.Add("authorization", "Bearer "+*token)

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error retrieving organization. Status:" + resp.Status + ". Body: " + string(body))
	}

	log.Trace().Str("body", string(body)).Msg("/svc/core/api/v1/organizations/mine response")

	var organizations []dto.OrganizationOutDTO
	err = json.Unmarshal(body, &organizations)
	if err != nil {
		return nil, err
	}

	if len(organizations) == 0 {
		return nil, errors.New("no organization found")
	}

	if len(organizations) > 1 {
		return nil, errors.New("more than one organization found")
	}

	return &organizations[0], nil
}
