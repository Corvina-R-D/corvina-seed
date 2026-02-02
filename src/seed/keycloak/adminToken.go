package keycloak

import (
	"context"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/utils"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func getKeycloakUserIdInOrganization(ctx context.Context) (string, error) {
	keycloakOrigin := ctx.Value(utils.KeycloakOrigin).(string)
	masterToken, err := MasterToken(ctx)
	if err != nil {
		return "", err
	}
	realm := ctx.Value(utils.UserRealm).(string)
	username := ctx.Value(utils.AdminUserKey).(string)

	headers := map[string]string{
		"Authorization": "Bearer " + *masterToken,
		"Content-type":  "application/json",
	}

	endpoint := fmt.Sprintf("%s/auth/admin/realms/%s/users?username=%s", keycloakOrigin, realm, username)

	log.Trace().Str("endpoint", endpoint).Msg("GetKeycloakUserIdInOrganization request")

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get user: %s. response body %s", resp.Status, string(body))
	}

	log.Trace().Str("body", string(body)).Msg("GetKeycloakUserIdInOrganization response")

	var response []map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response) == 0 {
		return "", fmt.Errorf("user not found")
	}

	userID, ok := response[0]["id"].(string)
	if !ok {
		return "", fmt.Errorf("user ID not found in response")
	}

	return userID, nil
}

func impersonateUser(ctx context.Context, realm string, userID string) (output *dto.OpenIdConnectTokenOutDTO, err error) {
	domain := ctx.Value(utils.DomainKey).(string)
	masterToken, err := MasterToken(ctx)
	if err != nil {
		return
	}
	keycloakOrigin := ctx.Value(utils.KeycloakOrigin).(string)
	headers := map[string]string{
		"Authorization": "Bearer " + *masterToken,
		"Content-type":  "application/json",
	}

	endpoint := fmt.Sprintf("%s/auth/admin/realms/%s/users/%s/impersonation", keycloakOrigin, realm, userID)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, nil)
	if err != nil {
		return
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to impersonate user: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	log.Trace().Str("body", string(body)).Msg("ImpersonateUser response")

	// Generate PKCE code verifier and challenge (S256)
	verifierBytes := make([]byte, 64)
	if _, err = rand.Read(verifierBytes); err != nil {
		return nil, err
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)
	sha := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(sha[:])

	params := url.Values{}
	params.Set("response_mode", "fragment")
	params.Set("response_type", "code")
	params.Set("client_id", "corvina-web-portal")
	params.Set("scope", fmt.Sprintf("openid org:%s", realm))
	params.Set("redirect_uri", fmt.Sprintf("https://%s.app.%s/", realm, domain))
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")

	endpoint = fmt.Sprintf("%s/auth/realms/%s/protocol/openid-connect/auth?%s", keycloakOrigin, realm, params.Encode())

	log.Debug().Str("endpoint", endpoint).Msg("ImpersonateUser request")

	req, err = http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return
	}

	for _, value := range resp.Cookies() {
		req.AddCookie(value)
	}

	resp, err = utils.HttpClientNoFollow.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")

	log.Trace().Str("location", location).Msg("ImpersonateUser response")

	query, err := url.ParseQuery(location)
	if err != nil {
		return nil, err
	}

	code := query.Get("code")

	log.Trace().Str("code", code).Msg("ImpersonateUser response")

	// Get access token from code
	endpoint = fmt.Sprintf("%s/auth/realms/%s/protocol/openid-connect/token", keycloakOrigin, realm)
	data := url.Values{}
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", "corvina-web-portal")
	data.Set("scope", fmt.Sprintf("openid org:%s", realm))
	data.Set("redirect_uri", fmt.Sprintf("https://%s.app.%s/", realm, domain))
	// include PKCE verifier for token exchange
	data.Set("code_verifier", codeVerifier)

	req, err = http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, value := range resp.Cookies() {
		req.AddCookie(value)
	}

	resp, err = utils.HttpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("failed to get access token: %s - body: %s", resp.Status, string(body))
	}

	if err := json.Unmarshal(body, &output); err != nil {
		return nil, err
	}

	return output, nil
}

func fetchAdminUserToken(ctx context.Context) (*dto.OpenIdConnectTokenOutDTO, error) {
	userId, err := getKeycloakUserIdInOrganization(ctx)
	if err != nil {
		return nil, err
	}

	token, err := impersonateUser(ctx, ctx.Value(utils.UserRealm).(string), userId)

	return token, err
}

var adminTokenCache *dto.OpenIdConnectTokenOutDTO

func AdminToken(ctx context.Context) (*string, error) {
	if adminTokenCache != nil {
		return &adminTokenCache.AccessToken, nil
	}

	token, err := fetchAdminUserToken(ctx)
	if err != nil {
		return nil, err
	}

	adminTokenCache = token

	duration := time.Duration(utils.Percent(token.ExpiresIn, 0.9)) * time.Second
	time.AfterFunc(duration, func() {
		adminTokenCache = nil
	})

	return &token.AccessToken, nil
}
