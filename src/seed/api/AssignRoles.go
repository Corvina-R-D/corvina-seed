package api

import (
	"context"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"corvina/corvina-seed/src/utils/int64s"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

func AssignRolesToUser(ctx context.Context, organizationId int64, userId int64, roles []int64) error {
	userGroupId, err := GetUserGroupIdFromUserId(ctx, organizationId, userId)
	if err != nil {
		return err
	}

	return AssignRolesToUserGroup(ctx, organizationId, userGroupId, roles)
}

func AssignRolesToUserGroup(ctx context.Context, organizationId int64, userGroupId int64, roles []int64) error {
	origin := ctx.Value(utils.OriginKey).(string)

	url := origin + fmt.Sprintf("/svc/core/api/v1/organizations/%d/userGroups/%d/roles/%s", organizationId, userGroupId, int64s.Join(roles, ","))
	log.Debug().Str("url", url).Msg("")
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("accept", "application/json")
	token, err := keycloak.AdminToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Add("authorization", "Bearer "+*token)

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("error assigning roles to user group, status=%s", resp.Status)
	}

	return nil
}
