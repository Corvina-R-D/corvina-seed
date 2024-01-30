package api

import (
	"bytes"
	"context"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type CreateMappingInDTO struct {
	Name string         `json:"name"`
	Data dto.IoTDataDTO `json:"data"`
}

type CreateMappingOutDTO struct {
	Id      string         `json:"id"`
	Name    string         `json:"name"`
	Version string         `json:"version"`
	Data    dto.IoTDataDTO `json:"json"`
	ModelId string         `json:"modelId"`
}

func CreateMapping(ctx context.Context, orgResourceId string, input CreateModelInDTO) (*CreateMappingOutDTO, error) {
	origin := ctx.Value(utils.OriginKey).(string)

	url := origin + "/svc/mappings/api/v1/presets?organization=" + orgResourceId

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(input)
	if err != nil {
		return nil, err
	}

	log.Trace().Msgf("CreateMapping: %s", buf.String())

	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, err
	}

	token, err := keycloak.AdminToken(ctx)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+*token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("error creating mapping %d %s", resp.StatusCode, string(body))
	}

	var out CreateMappingOutDTO
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}
