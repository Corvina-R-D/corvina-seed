package keycloak

import (
	"bytes"
	"context"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

var licenseTokenCache *dto.OpenIdConnectTokenOutDTO

func LicenseManagerToken(ctx context.Context) (*string, error) {
	if licenseTokenCache != nil {
		return &licenseTokenCache.AccessToken, nil
	}

	token, err := fetchLicenseManagerToken(ctx)
	if err != nil {
		return nil, err
	}

	licenseTokenCache = token

	duration := time.Duration(utils.Percent(token.ExpiresIn, 0.9)) * time.Second
	time.AfterFunc(duration, func() {
		licenseTokenCache = nil
	})

	return &token.AccessToken, nil
}

func fetchLicenseManagerToken(ctx context.Context) (*dto.OpenIdConnectTokenOutDTO, error) {
	keycloakOrigin := ctx.Value(utils.KeycloakOrigin).(string)
	username := ctx.Value(utils.LicenseManagerUser).(string)
	password := ctx.Value(utils.LicenseManagerPass).(string)
	endpoint := keycloakOrigin + "/auth/realms/master/protocol/openid-connect/token"

	formData := url.Values{}
	formData.Set("username", username)
	formData.Set("password", password)
	formData.Set("grant_type", "password")
	formData.Set("client_id", "license-manager")

	requestBody := formData.Encode()

	log.Trace().Str("endpoint", endpoint).Str("request body", requestBody).Msg("MasterToken request")

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBufferString(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, errors.New("error fetching license manager token. Status:" + resp.Status + ". Body: " + string(body))
	}

	log.Trace().Str("body", string(body)).Msg("LicenseManagerToken response")

	output := dto.OpenIdConnectTokenOutDTO{}

	if err := json.Unmarshal(body, &output); err != nil {
		return nil, err
	}

	return &output, nil
}
