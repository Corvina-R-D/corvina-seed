package api

import (
	"bytes"
	"context"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func CreateDeviceLicense(ctx context.Context) (activationKey *string, err error) {
	licenseManagerHost := ctx.Value(utils.LicenseHostKey).(string)

	endpoint := licenseManagerHost + "/api/v1/deviceLicenses"

	requestBody := struct {
		Notes         string `json:"notes"`
		ActivationKey string `json:"activationKey"`
	}{
		Notes:         "Created by corvina-seed",
		ActivationKey: uuid.New().String(),
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	if err = enc.Encode(requestBody); err != nil {
		return
	}

	log.Trace().Str("body", buf.String()).Msg("CreateDeviceLicense request")

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, &buf)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	token, err := keycloak.LicenseManagerToken(ctx)
	if err != nil {
		return
	}
	req.Header.Set("corvina-realm-token", *token)

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error creating device license %s %s", resp.Status, string(body))
	}

	var responseBody struct {
		ActivationKey string `json:"activationKey"`
	}

	if err = json.Unmarshal(body, &responseBody); err != nil {
		return
	}

	activationKey = &responseBody.ActivationKey

	return
}
