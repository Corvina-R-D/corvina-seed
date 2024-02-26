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
	"strconv"

	"github.com/rs/zerolog/log"
)

type CreateOrganizationInDTO struct {
	Name string `json:"name"`
}

type CreateOrganizationOutDTO struct {
	ID                        int64  `json:"id"`
	Name                      string `json:"name"`
	Label                     string `json:"label"`
	Status                    string `json:"status"`
	ResourceID                string `json:"resourceId"`
	PrivateAccess             bool   `json:"privateAccess"`
	AllowDisablePrivateAccess bool   `json:"allowDisablePrivateAccess"`
	HostnameAllowed           bool   `json:"hostnameAllowed"`
	DataEnabled               bool   `json:"dataEnabled"`
	VpnEnabled                bool   `json:"vpnEnabled"`
	VpnOtpRequired            bool   `json:"vpnOtpRequired"`
	UserCanAccess             bool   `json:"userCanAccess"`
	StoreEnabled              bool   `json:"storeEnabled"`
}

func CreateSubOrganization(ctx context.Context, orgId int64, input CreateOrganizationInDTO) (*CreateOrganizationOutDTO, error) {
	origin := ctx.Value(utils.OriginKey).(string)

	url := origin + "/svc/core/api/v1/organizations/" + strconv.FormatInt(orgId, 10) + "?addDefaultRoles=true&waitLicenseManager=true"

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	if err := enc.Encode(input); err != nil {
		return nil, err
	}

	log.Trace().Str("body", buf.String()).Msg("CreateSubOrganization request")

	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, errors.New("error creating sub organization. Status:" + resp.Status + ". Body: " + string(body))
	}

	log.Trace().Str("body", string(body)).Msg("CreateSubOrganization response")

	var output CreateOrganizationOutDTO
	if err := json.Unmarshal(body, &output); err != nil {
		return nil, err
	}

	return &output, nil

}
