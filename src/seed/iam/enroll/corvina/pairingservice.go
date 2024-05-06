package corvina

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

type LicenseData struct {
	PlatformPairingApiUrl string
	LogicalId             string
	InstanceId            string
	OrganizationId        string
	ApiKey                string
	Realm                 string
	BrokerUrls            []string `json:"-"`
	BrokerUrlsRaw         string   `json:"brokerUrls"`
}

type CrtData struct {
	Data struct {
		ClientCrt string `json:"client_crt"`
	} `json:"data"`
}

type PairingService struct {
	licenseData LicenseData
}

func NewPairingService(licenseData LicenseData) *PairingService {
	return &PairingService{
		licenseData: licenseData,
	}
}

func (p *PairingService) DoPairing(protocol string, csr string) (CrtData, error) {
	url := fmt.Sprintf("%s/devices/%s/protocols/%s/credentials", p.licenseData.PlatformPairingApiUrl, p.licenseData.LogicalId, protocol)
	reqBody, _ := json.Marshal(map[string]interface{}{
		"data": map[string]string{
			"csr": csr,
		},
	})
	log.Debug().Str("req body", string(reqBody)).Msg("")
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.licenseData.ApiKey)

	resp, err := Client.Do(req)
	if err != nil {
		return CrtData{}, err
	}
	defer resp.Body.Close()

	log.Info().Int("response status", resp.StatusCode).Msg("")

	var data CrtData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return CrtData{}, err
	}

	return data, nil
}

func (p *PairingService) Verify(protocol string, crt string) (bool, error) {
	url := fmt.Sprintf("%s/devices/%s/protocols/%s/credentials/verify", p.licenseData.PlatformPairingApiUrl, p.licenseData.LogicalId, protocol)
	reqBody, _ := json.Marshal(map[string]interface{}{
		"data": map[string]string{
			"client_crt": crt,
		},
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.licenseData.ApiKey)

	resp, err := Client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	log.Info().Int("response status", resp.StatusCode).Msg("")
	return resp.StatusCode == http.StatusOK, nil
}
