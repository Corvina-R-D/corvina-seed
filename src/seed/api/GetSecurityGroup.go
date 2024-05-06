package api

import (
	"context"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

type SecurityGroupOutDTO struct {
	ID            int64  `json:"id"`
	Label         string `json:"label"`
	HwID          string `json:"hwId"`
	OrgResourceID string `json:"orgResourceId"`
}

func GetSecurityPolicy(ctx context.Context, organizationId int64, name string) (*SecurityGroupOutDTO, error) {
	filter := SecurityPolicyFilterDTO{
		Name: name,
	}

	securityPolicies, err := GetSecurityPolicies(ctx, organizationId, filter)
	if err != nil {
		return nil, err
	}

	if len(*securityPolicies) == 0 {
		return nil, errors.New("security policy not found")
	}

	if len(*securityPolicies) > 1 {
		return nil, errors.New("more than one security policy found")
	}

	return &(*securityPolicies)[0], nil
}

type SecurityPolicyFilterDTO struct {
	Name string `json:"name"`
}

func GetSecurityPolicies(ctx context.Context, organizationId int64, filter SecurityPolicyFilterDTO) (*[]SecurityGroupOutDTO, error) {
	origin := ctx.Value(utils.OriginKey).(string)

	url := fmt.Sprintf("%s/svc/core/api/v1/organizations/%d/securityPolicies", origin, organizationId)
	log.Debug().Str("url", url).Msg("")

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if filter.Name != "" {
		q.Add("name", filter.Name)
	}
	req.URL.RawQuery = q.Encode()

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

	var out CorePagedDTO[SecurityGroupOutDTO]
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, err
	}

	return &out.Content, nil
}
