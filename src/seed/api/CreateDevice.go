package api

import (
	"bytes"
	"context"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"corvina/corvina-seed/src/utils/ref"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func CreateDevice(ctx context.Context, orgResourceId string, name string) (activationKey *string, err error) {

	activationKey, err = CreateDeviceLicense(ctx)
	if err != nil {
		return
	}
	log.Info().Str("activationKey", *activationKey).Msg("Device license created")

	err = activateDeviceLicense(ctx, *activationKey, name, orgResourceId)
	if err != nil {
		return nil, err
	}

	log.Info().Str("device name", name).Msg("Device license activated and device created")

	eachDeviceHasMapping := utils.CtxValueOrDefault(ctx, utils.EachDeviceHasMapping, false)

	if !eachDeviceHasMapping {
		return
	}

	model, err := CreateRandomModel(ctx, orgResourceId)
	if err != nil {
		return
	}

	data := model.Data
	for key, prop := range data.Properties {
		prop.Mode = ref.String("R")
		prop.SendPolicy = &dto.SendPolicyDTO{
			Triggers: []dto.TriggerDTO{
				{
					ChangeMask:        "value",
					MinIntervalMs:     1000,
					SkipFirstNChanges: 0,
					Type:              "onchange",
				},
			},
		}
		prop.Datalink = &dto.DatalinkDTO{
			Source: key,
		}
		prop.HistoryPolicy = &dto.HistoryPolicyDTO{
			Enabled: true,
		}
		prop.Version = ref.String("1.0.0")
		data.Properties[key] = prop
	}
	mapping, err := CreateMapping(ctx, orgResourceId, CreateModelInDTO{
		Name: utils.RandomName(),
		Data: data,
	})
	if err != nil {
		return
	}

	log.Info().Interface("mapping", mapping).Msg("Mapping created")

	device, err := TryGetDeviceNTimes(ctx, orgResourceId, name, uint8(20))
	if err != nil {
		return
	}

	log.Debug().Interface("device", device).Msg("Device found in mapping service")

	err = SetDeviceMapping(ctx, &mapping.Id, &device.DeviceID)
	if err != nil {
		return
	}

	return
}

func TryGetDeviceNTimes(ctx context.Context, orgResourceId string, name string, n uint8) (device *DeviceMappingServiceOutDTO, err error) {
	for i := uint8(0); i < n; i++ {
		device, err = GetDeviceFromMappingService(ctx, orgResourceId, name)
		if err != nil {
			time.Sleep(1 * time.Second)
			log.Info().Uint8("i", i).Err(err).Msg("Device not found")
			continue
		}

		log.Debug().Uint8("i", i).Interface("device", device).Msg("Device found")

		return
	}

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

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return
	}
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
