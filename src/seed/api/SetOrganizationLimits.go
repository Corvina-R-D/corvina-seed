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

	"github.com/rs/zerolog/log"
)

type SetLimitInDTO struct {
	OrgResourceID string `json:"orgResourceId"`
	ResourceType  string `json:"resourceType"`
	Quantity      string `json:"quantity"`
	Valid         bool   `json:"valid"`
}

func SetNumberOfOrgLimit(ctx context.Context, orgResourceID string, quantity string) (err error) {
	return SetOrganizationLimit(ctx, &SetLimitInDTO{
		OrgResourceID: orgResourceID,
		ResourceType:  "ORGANIZATIONS",
		Quantity:      quantity,
		Valid:         false,
	})
}

func SetNumberOfUsersLimit(ctx context.Context, orgResourceID string, quantity string) (err error) {
	return SetOrganizationLimit(ctx, &SetLimitInDTO{
		OrgResourceID: orgResourceID,
		ResourceType:  "USERS",
		Quantity:      quantity,
		Valid:         false,
	})
}

func SetNumberOfDevicesLimit(ctx context.Context, orgResourceID string, quantity string) (err error) {
	return SetOrganizationLimit(ctx, &SetLimitInDTO{
		OrgResourceID: orgResourceID,
		ResourceType:  "DEVICES",
		Quantity:      quantity,
		Valid:         false,
	})
}

func SetNumberOfDeviceDataLimit(ctx context.Context, orgResourceID string, quantity string) (err error) {
	return SetOrganizationLimit(ctx, &SetLimitInDTO{
		OrgResourceID: orgResourceID,
		ResourceType:  "DEVICE_DATA",
		Quantity:      quantity,
		Valid:         false,
	})
}

func SetNumberOfDeviceVpnLimit(ctx context.Context, orgResourceID string, quantity string) (err error) {
	return SetOrganizationLimit(ctx, &SetLimitInDTO{
		OrgResourceID: orgResourceID,
		ResourceType:  "DEVICE_VPN",
		Quantity:      quantity,
		Valid:         false,
	})
}

func SetAllLimitToUnlimited(ctx context.Context, orgResourceID string) (err error) {
	err = SetNumberOfOrgLimit(ctx, orgResourceID, "-1")
	if err != nil {
		return
	}
	err = SetNumberOfUsersLimit(ctx, orgResourceID, "-1")
	if err != nil {
		return
	}
	err = SetNumberOfDevicesLimit(ctx, orgResourceID, "-1")
	if err != nil {
		return
	}
	err = SetNumberOfDeviceDataLimit(ctx, orgResourceID, "-1")
	if err != nil {
		return
	}
	err = SetNumberOfDeviceVpnLimit(ctx, orgResourceID, "-1")
	if err != nil {
		return
	}
	return
}

func SetOrganizationLimit(ctx context.Context, input *SetLimitInDTO) (err error) {
	origin := ctx.Value(utils.OriginKey).(string)

	url := origin + "/svc/license/api/v1/limits"

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	if err = enc.Encode(input); err != nil {
		return
	}

	log.Trace().Str("body", buf.String()).Msg("SetOrganizationLimit request")

	req, err := http.NewRequestWithContext(ctx, "PUT", url, &buf)
	if err != nil {
		return err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	token, err := keycloak.AdminToken(ctx)
	if err != nil {
		return
	}
	req.Header.Add("authorization", "Bearer "+*token)

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return errors.New("error setting limit. Status:" + resp.Status + ". Body: " + string(body))
	}

	log.Trace().Str("body", string(body)).Msg("SetOrganizationLimit response")

	return
}
