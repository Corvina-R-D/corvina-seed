package api

import (
	"context"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

func AssignSecurityPolicyToUser(ctx context.Context, organizationId int64, deviceId int64, userId int64) error {
	origin := ctx.Value(utils.OriginKey).(string)

	url := origin + fmt.Sprintf("/svc/core/api/v1/organizations/%d/users/%d/securityPolicies/%d", organizationId, userId, deviceId)
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("error assigning device to user, status=%s body=%s", resp.Status, resp.Body)
	}

	return nil
}
