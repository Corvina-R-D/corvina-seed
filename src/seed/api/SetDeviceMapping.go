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
)

func SetDeviceMapping(ctx context.Context, mappingId *string, deviceId *string) (err error) {
	origin := ctx.Value(utils.OriginKey).(string)

	endpoint := origin + "/svc/mappings/api/v1/devices/" + *deviceId

	requestBody := map[string]string{
		"presetId": *mappingId,
	}
	jsonPayload, err := json.Marshal(requestBody)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return
	}

	token, err := keycloak.AdminToken(ctx)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", "Bearer "+*token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("error setting device mapping. Status: %d %s", resp.StatusCode, string(body))

	}

	return
}
