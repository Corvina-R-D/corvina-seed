package api

import (
	"bytes"
	"context"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

func CreateDevice(ctx context.Context, orgResourceId string, name string) (err error) {

	activationKey, err := CreateDeviceLicense(ctx)
	if err != nil {
		return
	}
	log.Info().Str("activationKey", *activationKey).Msg("Device license created")

	err = activateDeviceLicense(ctx, *activationKey, name, orgResourceId)
	if err != nil {
		return err
	}

	log.Info().Str("device name", name).Msg("Device license activated and device created")

	return
}

func activateDeviceLicense(ctx context.Context, activationKey string, deviceAlias string, orgResourceId string) (err error) {
	licenseManagerHost := ctx.Value(utils.LicenseHostKey).(string)
	url := fmt.Sprintf("%s/api/v1/deviceLicenses/activate", licenseManagerHost)

	payload := map[string]string{
		"alias":         deviceAlias,
		"activationKey": activationKey,
		"orgResourceId": orgResourceId,
	}

	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	token, err := keycloak.AdminToken(ctx)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", *token)

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Error().Str("status", resp.Status).Msg("Error activating device license")
		return
	}

	return
}
