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

var masterTokenCache *dto.OpenIdConnectTokenOutDTO

func MasterToken(ctx context.Context) (*string, error) {
	if masterTokenCache != nil {
		return &masterTokenCache.AccessToken, nil
	}

	token, err := fetchMasterToken(ctx)
	if err != nil {
		return nil, err
	}

	masterTokenCache = token

	duration := time.Duration(utils.Percent(token.ExpiresIn, 0.9)) * time.Second
	time.AfterFunc(duration, func() {
		masterTokenCache = nil
	})

	return &token.AccessToken, nil
}

func fetchMasterToken(ctx context.Context) (*dto.OpenIdConnectTokenOutDTO, error) {
	keycloakOrigin := ctx.Value(utils.KeycloakOrigin).(string)
	masterClientId := ctx.Value(utils.KeycloakMasterClientId).(string)
	masterClientSecret := ctx.Value(utils.KeycloakMasterClientSecret).(string)

	endpoint := keycloakOrigin + "/auth/realms/master/protocol/openid-connect/token"

	formData := url.Values{}
	formData.Set("client_id", masterClientId)
	formData.Set("client_secret", masterClientSecret)
	formData.Set("grant_type", "client_credentials")

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
		return nil, errors.New("error fetching master token. Status:" + resp.Status + ". Body: " + string(body))
	}

	log.Trace().Str("body", string(body)).Msg("MasterToken response")

	output := dto.OpenIdConnectTokenOutDTO{}

	if err := json.Unmarshal(body, &output); err != nil {
		return nil, err
	}

	return &output, nil
}
