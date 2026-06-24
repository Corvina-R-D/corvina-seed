package api

import (
	"context"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/seed/keycloak"
	"corvina/corvina-seed/src/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

// GetOrganizationMine resolves the organization to work in.
//
// /organizations/mine only returns the root organizations the user belongs to;
// sub organizations are not listed there. When --org-resource-id points to a
// sub organization (e.g. "exor.testdavide3", or just "testdavide3" when there is
// a single root), this descends the organization tree to resolve it.
func GetOrganizationMine(ctx context.Context) (*dto.OrganizationOutDTO, error) {
	orgResourceId := ctx.Value(utils.OrgResourceId).(string)

	mine, err := fetchOrganizationsMine(ctx)
	if err != nil {
		return nil, err
	}

	if len(mine) == 0 {
		return nil, errors.New("no organization found")
	}

	if orgResourceId == "" {
		if len(mine) == 1 {
			return &mine[0], nil
		}
		return nil, errors.New("more than one organization found and no --org-resource-id provided")
	}

	// Exact match among the root organizations.
	for i := range mine {
		if mine[i].ResourceID == orgResourceId {
			return &mine[i], nil
		}
	}

	// Otherwise descend the tree to resolve a sub organization.
	return resolveSubOrganization(ctx, mine, orgResourceId)
}

func fetchOrganizationsMine(ctx context.Context) ([]dto.OrganizationOutDTO, error) {
	origin := ctx.Value(utils.OriginKey).(string)

	url := origin + "/svc/core/api/v1/organizations/mine"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error retrieving organization. Status:" + resp.Status + ". Body: " + string(body))
	}

	log.Trace().Str("body", string(body)).Msg("/svc/core/api/v1/organizations/mine response")

	var organizations []dto.OrganizationOutDTO
	if err = json.Unmarshal(body, &organizations); err != nil {
		return nil, err
	}

	return organizations, nil
}

// resolveSubOrganization walks the organization tree following the dotted,
// hierarchical resource id (e.g. "exor.testdavide3") starting from the matching
// root organization. A short id without dots is auto-prefixed with the single
// root organization resource id.
func resolveSubOrganization(ctx context.Context, mine []dto.OrganizationOutDTO, orgResourceId string) (*dto.OrganizationOutDTO, error) {
	target := orgResourceId
	if !strings.Contains(target, ".") {
		if len(mine) != 1 {
			return nil, fmt.Errorf("organization %q not found among your root organizations; you have multiple root organizations, pass the full hierarchical resource id (e.g. <root>.%s)", orgResourceId, orgResourceId)
		}
		target = mine[0].ResourceID + "." + target
	}

	segments := strings.Split(target, ".")

	var cur *dto.OrganizationOutDTO
	for i := range mine {
		if mine[i].ResourceID == segments[0] {
			cur = &mine[i]
			break
		}
	}
	if cur == nil {
		return nil, fmt.Errorf("root organization %q not found among your organizations", segments[0])
	}

	cumulative := segments[0]
	for _, seg := range segments[1:] {
		cumulative = cumulative + "." + seg
		child, err := findChildByResourceId(ctx, cur.ID, cumulative)
		if err != nil {
			return nil, err
		}
		cur = child
	}

	return cur, nil
}

// findChildByResourceId pages through the direct sub organizations of parentId
// and returns the one whose resource id matches.
func findChildByResourceId(ctx context.Context, parentId int64, resourceId string) (*dto.OrganizationOutDTO, error) {
	origin := ctx.Value(utils.OriginKey).(string)

	for page := 0; ; page++ {
		url := fmt.Sprintf("%s/svc/core/api/v1/organizations/%d/organizations?page=%d", origin, parentId, page)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

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
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error listing sub organizations of %d. Status: %s. Body: %s", parentId, resp.Status, string(body))
		}

		var out CorePagedDTO[dto.OrganizationOutDTO]
		if err = json.Unmarshal(body, &out); err != nil {
			return nil, err
		}

		for i := range out.Content {
			if out.Content[i].ResourceID == resourceId {
				return &out.Content[i], nil
			}
		}

		if out.Last || len(out.Content) == 0 || page >= out.TotalPages-1 {
			break
		}
	}

	return nil, fmt.Errorf("sub organization with resource id %q not found under organization %d", resourceId, parentId)
}
