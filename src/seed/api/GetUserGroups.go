package api

import (
	"context"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

type UserGroupFilterDTO struct {
	UserId int64  `json:"userId"`
	Type   string `json:"type"`
}

type UserGroupDTO struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	OrganizationID int64  `json:"organizationId"`
	Type           string `json:"type"`
	Owner          string `json:"owner"`
}

func GetUserGroups(ctx context.Context, organizationId int64, filter UserGroupFilterDTO) ([]UserGroupDTO, error) {
	origin := ctx.Value(utils.OriginKey).(string)

	url := fmt.Sprintf("%s/svc/core/api/v1/organizations/%d/userGroups", origin, organizationId)
	log.Debug().Str("url", url).Msg("")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("userId", strconv.FormatInt(filter.UserId, 10))
	q.Add("type", filter.Type)
	req.URL.RawQuery = q.Encode()

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

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("error getting user groups, status=%d body=%s", resp.StatusCode, string(body))
	}

	var out CorePagedDTO[UserGroupDTO]
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, err
	}

	return out.Content, nil
}

func GetUserGroupIdFromUserId(ctx context.Context, organizationId int64, userId int64) (int64, error) {
	userGroups, err := GetUserGroups(ctx, organizationId, UserGroupFilterDTO{
		UserId: userId,
		Type:   "SELF_USER_ANY",
	})
	if err != nil {
		return 0, err
	}

	if len(userGroups) == 0 {
		return 0, fmt.Errorf("user %d not found in any group", userId)
	}

	return userGroups[0].ID, nil
}
