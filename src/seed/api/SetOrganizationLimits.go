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
	"time"

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

// SetAllLimitToUnlimited imposta tutti i limiti a -1 (illimitato).
// delay viene applicato prima di ogni singola PUT: con chiamate ravvicinate
// anche i 5 limit-call rispettano il rate limiter del server.
func SetAllLimitToUnlimited(ctx context.Context, orgResourceID string, delay time.Duration) (err error) {
	calls := []func() error{
		func() error { return SetNumberOfOrgLimit(ctx, orgResourceID, "-1") },
		func() error { return SetNumberOfUsersLimit(ctx, orgResourceID, "-1") },
		func() error { return SetNumberOfDevicesLimit(ctx, orgResourceID, "-1") },
		func() error { return SetNumberOfDeviceDataLimit(ctx, orgResourceID, "-1") },
		func() error { return SetNumberOfDeviceVpnLimit(ctx, orgResourceID, "-1") },
	}
	for _, call := range calls {
		time.Sleep(delay)
		if err = call(); err != nil {
			return
		}
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
