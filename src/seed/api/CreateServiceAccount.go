package api

import (
	"bytes"
	"context"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CreateUserInDTO struct {
	Email                          string   `json:"email"`
	Username                       string   `json:"username"`
	EmailVerified                  bool     `json:"emailVerified"`
	TemporaryPassword              bool     `json:"temporaryPassword"`
	ServiceAccount                 bool     `json:"serviceAccount"`
	ServiceAccountClientWebOrigins []string `json:"serviceAccountClientWebOrigins"`
	PasswordChangeInvitation       bool     `json:"passwordChangeInvitation"`
	GroupPoliciesEnabled           bool     `json:"groupPoliciesEnabled"`
	MemberOf                       []int64  `json:"memberOf"`
	RegularUserOf                  []int64  `json:"regularUserOf"`
	AdminUserOf                    []int64  `json:"adminUserOf"`
}

type UserDTO struct {
	ID                   int    `json:"id"`
	Username             string `json:"username"`
	Email                string `json:"email"`
	FirstName            any    `json:"firstName"`
	LastName             any    `json:"lastName"`
	Country              any    `json:"country"`
	ServiceAccount       bool   `json:"serviceAccount"`
	ServiceAccountSecret string `json:"serviceAccountSecret"`
	Owner                string `json:"owner"`
	OwnerRef             string `json:"ownerRef"`
	GroupPoliciesEnabled bool   `json:"groupPoliciesEnabled"`
}

func CreateServiceAccount(ctx context.Context, organizationId int64, name string) (user *UserDTO, err error) {
	userInDTO := CreateUserInDTO{
		Username:                 name,
		Email:                    "--",
		ServiceAccount:           true,
		EmailVerified:            true,
		GroupPoliciesEnabled:     false,
		MemberOf:                 []int64{},
		RegularUserOf:            []int64{},
		AdminUserOf:              []int64{},
		TemporaryPassword:        true,
		PasswordChangeInvitation: true,
	}

	return CreateUser(ctx, organizationId, userInDTO)
}

func CreateUser(ctx context.Context, organizationId int64, userInDTO CreateUserInDTO) (*UserDTO, error) {
	origin := ctx.Value(utils.OriginKey).(string)

	url := origin + fmt.Sprintf("/svc/core/api/v1/organizations/%d/users", organizationId)

	jsonPayload, err := json.Marshal(userInDTO)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("error creating user, status=%d body=%s", resp.StatusCode, string(body))
	}

	var user UserDTO
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
