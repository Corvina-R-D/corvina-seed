package api

import (
	"bytes"
	"context"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

type CreateDeviceGroupInDTO struct {
	Name string `json:"name"`
}

func CreateDeviceGroup(ctx context.Context, organizationId int64, input CreateDeviceGroupInDTO) error {
	origin := ctx.Value(utils.OriginKey).(string)

	url := origin + "/svc/core/api/v1/organizations/" + strconv.FormatInt(organizationId, 10) + "/securityPolicies"

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	if err := enc.Encode(input); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	token, err := keycloak.AdminToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Add("authorization", "Bearer "+*token)

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return errors.New("error creating device group. Status:" + resp.Status + ". Body: " + string(body))
	}

	log.Trace().Str("body", string(body)).Msg("CreateDeviceGroup response")

	return nil
}
