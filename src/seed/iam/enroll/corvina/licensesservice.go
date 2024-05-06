package corvina

import (
	"encoding/json"
	"fmt"
	"strings"
)

type LicensesService struct {
	pairingService  *PairingService
	activationKey   string
	pairingEndpoint string
}

const protocol = "corvina_mqtt_v1"

func NewLicensesService(pairingEndpoint string, activationKey string) *LicensesService {
	return &LicensesService{
		pairingEndpoint: pairingEndpoint,
		activationKey:   activationKey,
	}
}

func (l *LicensesService) Init() (LicenseData, error) {
	resp, err := Client.Get(fmt.Sprintf("%s?activationKey=%s&serialNumber=", l.pairingEndpoint, l.activationKey))
	if err != nil {
		return LicenseData{}, err
	}
	defer resp.Body.Close()

	instanceId := resp.Header.Get("x-instance-id")
	organizationId := resp.Header.Get("x-organization-id")

	var data LicenseData
	err = json.NewDecoder(resp.Body).Decode(&data)
	data.InstanceId = instanceId
	data.OrganizationId = organizationId
	if err != nil {
		return LicenseData{}, err
	}

	data.BrokerUrls = strings.Split(data.BrokerUrlsRaw, ",")

	l.pairingService = NewPairingService(data)

	if data.BrokerUrls != nil {
		for i, url := range data.BrokerUrls {
			if strings.HasPrefix(url, protocol) {
				data.BrokerUrls[i] = strings.Replace(url, protocol, "mqtts", 1)
			}
		}
	}

	return data, nil
}

func (l *LicensesService) DoPairing(csr string) (CrtData, error) {
	return l.pairingService.DoPairing(protocol, csr)
}

func (l *LicensesService) Verify(crt string) (bool, error) {
	return l.pairingService.Verify(protocol, crt)
}
