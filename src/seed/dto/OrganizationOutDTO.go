package dto

type OrganizationOutDTO struct {
	ID                        int      `json:"id"`
	Name                      string   `json:"name"`
	Label                     string   `json:"label"`
	Status                    string   `json:"status"`
	ResourceID                string   `json:"resourceId"`
	PrivateAccess             bool     `json:"privateAccess"`
	AllowDisablePrivateAccess bool     `json:"allowDisablePrivateAccess"`
	Hostname                  string   `json:"hostname"`
	HostnameAllowed           bool     `json:"hostnameAllowed"`
	DataEnabled               bool     `json:"dataEnabled"`
	VpnEnabled                bool     `json:"vpnEnabled"`
	VpnPairingMode            string   `json:"vpnPairingMode"`
	VpnOtpRequired            bool     `json:"vpnOtpRequired"`
	IPAddressesWhitelist      []string `json:"ipAddressesWhitelist"`
	UserCanAccess             bool     `json:"userCanAccess"`
	StoreEnabled              bool     `json:"storeEnabled"`
}
