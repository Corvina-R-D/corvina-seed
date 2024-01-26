package api

import (
	"bytes"
	"context"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type CreateModelInDataPropertiesDTO struct {
	Type string `json:"type"`
}

type CreateModelInDataDTO struct {
	Type        string                                    `json:"type"`
	InstanceOf  string                                    `json:"instanceOf"`
	Properties  map[string]CreateModelInDataPropertiesDTO `json:"properties"`
	Label       string                                    `json:"label"`
	Unit        string                                    `json:"unit"`
	Description string                                    `json:"description"`
	Tags        []string                                  `json:"tags"`
}

type CreateModelInDTO struct {
	Name string               `json:"name"`
	Data CreateModelInDataDTO `json:"data"`
}

func CreateModel(ctx context.Context, orgResourceId string, input CreateModelInDTO) error {
	origin := ctx.Value(utils.OriginKey).(string)
	apiKey := ctx.Value(utils.ApiKey).(string)

	url := origin + "/svc/mappings/api/v1/models?organization=" + orgResourceId

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	if err := enc.Encode(input); err != nil {
		return err
	}

	log.Trace().Str("body", buf.String()).Msg("CreateModel request")

	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("X-Api-Key", apiKey)

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return errors.New("error creating model. Status:" + resp.Status + ". Body: " + string(body))
	}

	log.Trace().Str("body", string(body)).Msg("CreateModel response")

	return nil
}
