package api

import (
	"context"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type RoleOutDTO struct {
	ID                      int64  `json:"id"`
	Name                    string `json:"name"`
	Label                   string `json:"label"`
	Description             string `json:"description"`
	Type                    string `json:"type"`
	Owner                   string `json:"owner"`
	OwnerRef                string `json:"ownerRef"`
	Enabled                 bool   `json:"enabled"`
	Deleted                 bool   `json:"deleted"`
	CreatedAt               int64  `json:"createdAt"`
	UpdatedAt               int64  `json:"updatedAt"`
	OrganizationID          int64  `json:"organizationId"`
	DeviceGeneralPermission any    `json:"deviceGeneralPermission"`
	VpnGeneralPermission    any    `json:"vpnGeneralPermission"`
	EnableAccessToApp       bool   `json:"enableAccessToApp"`
	OrgResourceID           string `json:"orgResourceId"`
}

type CorePagedDTO[T any] struct {
	Number        int  `json:"number"`
	Content       []T  `json:"content"`
	TotalElements int  `json:"totalElements"`
	TotalPages    int  `json:"totalPages"`
	Last          bool `json:"last"`
}

func GetFirstAdminApplicationRole(ctx context.Context, orgId int64) (role *RoleOutDTO, err error) {
	roles, err := GetApplicationRoles(ctx, orgId)
	if err != nil {
		return
	}

	for _, r := range *roles {
		if strings.Contains(r.Label, "Administrator") {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("admin role not found")
}

func GetApplicationRoles(ctx context.Context, orgId int64) (roles *[]RoleOutDTO, err error) {
	return GetRoles(ctx, orgId, []string{"APPLICATION"}, []string{"SYSTEM", "ORGANIZATION"})
}

func GetFirstAdminDeviceRole(ctx context.Context, orgId int64) (role *RoleOutDTO, err error) {
	roles, err := GetDeviceRoles(ctx, orgId)
	if err != nil {
		return
	}

	for _, r := range *roles {
		if strings.Contains(r.Label, "Administrator") {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("admin role not found")
}

func GetDeviceRoles(ctx context.Context, orgId int64) (roles *[]RoleOutDTO, err error) {
	return GetRoles(ctx, orgId, []string{"DEVICE"}, []string{"SYSTEM", "ORGANIZATION"})
}

func GetAppsSharingRoles(ctx context.Context, orgId int64) (roles *[]RoleOutDTO, err error) {
	return GetRoles(ctx, orgId, []string{"APPLICATION"}, []string{"APPLICATION"})
}

func GetRoles(ctx context.Context, orgId int64, types []string, owners []string) (roles *[]RoleOutDTO, err error) {
	origin := ctx.Value(utils.OriginKey).(string)

	endpoint := origin + "/svc/core/api/v1/organizations/" + fmt.Sprintf("%d", orgId) + "/roles?page=0&pageSize=500"

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

	var out CorePagedDTO[RoleOutDTO]
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return
	}

	log.Debug().Interface("out", out).Msg("getRoles")

	return &out.Content, nil

}
