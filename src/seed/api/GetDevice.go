package api

import (
	"context"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type DeviceMappingServiceOutDTO struct {
	Connected            bool   `json:"connected"`
	OrgResourceID        string `json:"orgResourceId"`
	RealmID              string `json:"realmId"`
	Deleted              bool   `json:"deleted"`
	ConfigurationApplied bool   `json:"configurationApplied"`
	ConfigurationSent    bool   `json:"configurationSent"`
	ID                   string `json:"id"`
	Label                string `json:"label"`
	CreationDate         int64  `json:"creationDate"`
	DeviceID             string `json:"deviceId"`
	UpdatedAt            int64  `json:"updatedAt"`
}

func GetDeviceFromMappingService(ctx context.Context, orgResourceId string, deviceName string) (device *DeviceMappingServiceOutDTO, err error) {
	origin := ctx.Value(utils.OriginKey).(string)

	endpoint := origin + "/svc/mappings/api/v1/devices?organization=" + orgResourceId + "&search=" + deviceName

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
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

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("error retrieving device %d %s", resp.StatusCode, string(body))
	}

	var out MappingPagedDTO[DeviceMappingServiceOutDTO]
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return
	}

	log.Debug().Interface("out", out).Msg("getDeviceFromMapping")

	if out.TotalElements == 0 {
		return nil, errors.New("unable to find any device with name " + deviceName + ". Zero TotalElements")
	}

	device = &out.Data[0]

	if device.DeviceID == "" {
		return nil, errors.New("unable to find any device with name " + deviceName + ". Empty DeviceID")
	}

	if device.Label != deviceName {
		return nil, errors.New("unable to find any device with name " + deviceName + ". Label mismatch")
	}

	return

}

type DeviceCoreServiceOutDTO struct {
	ID            int64  `json:"id"`
	Label         string `json:"label"`
	HwID          string `json:"hwId"`
	OrgResourceID string `json:"orgResourceId"`
}

func GetDeviceFromCoreService(ctx context.Context, organizationId int64, name string) (*DeviceCoreServiceOutDTO, error) {
	origin := ctx.Value(utils.OriginKey).(string)

	url := fmt.Sprintf("%s/svc/core/api/v1/organizations/%d/devices?deviceLabel=%s", origin, organizationId, name)
	log.Debug().Str("url", url).Msg("")
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	token, err := keycloak.AdminToken(ctx)
	if err != nil {
		return nil, err
	}
	req.Header.Add("authorization", "Bearer "+*token)

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting device, status=%s", resp.Status)
	}

	var out CorePagedDTO[DeviceCoreServiceOutDTO]
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, err
	}

	if out.TotalElements == 0 {
		return nil, errors.New("unable to find any device with name " + name + ". Zero TotalElements")
	}

	device := out.Content[0]

	if device.Label != name {
		return nil, errors.New("unable to find any device with name " + name + ". Label mismatch")
	}

	return &device, nil
}
