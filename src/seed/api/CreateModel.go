package api

import (
	"bytes"
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

type CreateModelInDTO struct {
	Name string         `json:"name"`
	Data dto.IoTDataDTO `json:"data"`
}

type CreateModelOutDTO struct {
	Id      string         `json:"id"`
	Name    string         `json:"name"`
	Version string         `json:"version"`
	Data    dto.IoTDataDTO `json:"json"`
}

func CreateRandomModel(ctx context.Context, orgResourceId string) (*CreateModelOutDTO, error) {
	name := utils.RandomName() + ":1.0.0"
	return CreateModel(ctx, orgResourceId, CreateModelInDTO{
		Name: name,
		Data: dto.IoTDataDTO{
			Type:       "object",
			InstanceOf: name,
			Properties: map[string]dto.IoTDataPropertiesDTO{
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
			Label:       utils.RandomName(),
			Unit:        "°C",
			Description: utils.RandomName(),
			Tags:        []string{utils.RandomName(), utils.RandomName()},
		},
	})
}

func CreateModel(ctx context.Context, orgResourceId string, input CreateModelInDTO) (*CreateModelOutDTO, error) {
	origin := ctx.Value(utils.OriginKey).(string)

	url := origin + "/svc/mappings/api/v1/models?organization=" + orgResourceId

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	if err := enc.Encode(input); err != nil {
		return nil, err
	}

	log.Trace().Str("body", buf.String()).Msg("CreateModel request")

	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, errors.New("error creating model. Status:" + resp.Status + ". Body: " + string(body))
	}

	log.Trace().Str("body", string(body)).Msg("CreateModel response")

	var output CreateModelOutDTO
	if err := json.Unmarshal(body, &output); err != nil {
		return nil, err
	}

	return &output, nil
}
