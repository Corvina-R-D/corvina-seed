package seed

import (
	"context"
	"corvina/corvina-seed/src/seed/api"
	"corvina/corvina-seed/src/seed/dto"
	"corvina/corvina-seed/src/utils"
	"fmt"
	"math/rand"

	"github.com/rs/zerolog/log"
)

// orgNode is a created organization we can later place devices into.
type orgNode struct {
	id         int64
	resourceID string
	depth      int64
}

// ExecuteRandomDistribution builds a nested, randomly shaped tree of
// organizations under the given root organization and then scatters the
// requested number of devices randomly across the whole tree (root included).
//
// Tree shape:
//   - level 1 has exactly input.OrganizationCount organizations under the root;
//   - every organization with depth < input.OrganizationTreeDepth gets a random
//     number of children in [0, input.MaxChildren], so the tree is irregular,
//     grows at most input.OrganizationTreeDepth levels deep and stays linear in
//     size instead of exploding as OrganizationCount^OrganizationTreeDepth.
func ExecuteRandomDistribution(ctx context.Context, input *dto.ExecuteInDTO, root *dto.OrganizationOutDTO) error {
	created := make([]orgNode, 0)

	// The root organization needs its ORGANIZATIONS/DEVICES resources available to
	// host sub organizations and devices, otherwise the API rejects creation with
	// "resource not available" (code 13010). New sub organizations get the same
	// treatment when created (see createSubOrgTolerant).
	if err := api.SetAllLimitToUnlimited(ctx, root.ResourceID); err != nil {
		log.Warn().Err(err).Str("orgResourceId", root.ResourceID).Msg("Could not set root organization limits to unlimited, continuing")
	}

	if input.OrganizationCount > 0 {
		utils.PrintlnBlue(fmt.Sprintf(
			"Building random organization tree: %d top-level orgs, max depth %d, under org %s",
			input.OrganizationCount, input.OrganizationTreeDepth, root.ResourceID,
		))

		nodes, err := buildRandomOrgTree(ctx, root.ID, 1, input.OrganizationCount, input.MaxChildren, input.OrganizationTreeDepth)
		if err != nil {
			return err
		}
		created = nodes

		utils.PrintlnGreen(fmt.Sprintf("Organizations created: %d", len(created)))
	}

	if input.DeviceCount > 0 {
		// Devices can land in any created organization or in the root itself.
		targets := make([]string, 0, len(created)+1)
		targets = append(targets, root.ResourceID)
		for _, node := range created {
			targets = append(targets, node.resourceID)
		}

		if err := distributeDevices(ctx, input.DeviceCount, targets); err != nil {
			return err
		}

		utils.PrintlnGreen(fmt.Sprintf("Devices created and distributed: %d over %d organizations", input.DeviceCount, len(targets)))
	}

	return nil
}

// buildRandomOrgTree creates `count` organizations under parentId at the given
// depth and recurses with a random branching factor of [0, maxChildren] until
// maxDepth is reached.
func buildRandomOrgTree(ctx context.Context, parentId int64, depth int64, count int64, maxChildren int64, maxDepth int64) ([]orgNode, error) {
	created := make([]orgNode, 0)

	for i := int64(0); i < count; i++ {
		node, err := createSubOrgTolerant(ctx, parentId, depth)
		if err != nil {
			return nil, err
		}
		created = append(created, *node)

		if depth >= maxDepth || maxChildren <= 0 {
			continue
		}

		// Random number of children in [0, maxChildren] for an irregular tree.
		childCount := rand.Int63n(maxChildren + 1)
		if childCount == 0 {
			continue
		}

		children, err := buildRandomOrgTree(ctx, node.id, depth+1, childCount, maxChildren, maxDepth)
		if err != nil {
			return nil, err
		}
		created = append(created, children...)
	}

	return created, nil
}

// createSubOrgTolerant creates a sub organization and tries to lift its limits.
// Unlike createSubOrg, a failure while setting limits is logged but does not
// abort the whole run, so a single limit hiccup does not waste a large bulk
// seeding.
func createSubOrgTolerant(ctx context.Context, parentId int64, depth int64) (*orgNode, error) {
	prefix := utils.CtxValueOrDefault(ctx, utils.OrgNamePrefix, "")
	subOrganization, err := api.CreateSubOrganization(ctx, parentId, api.CreateOrganizationInDTO{
		Name: prefix + utils.RandomName(),
	})
	if err != nil {
		return nil, err
	}

	if err := api.SetAllLimitToUnlimited(ctx, subOrganization.ResourceID); err != nil {
		log.Warn().Err(err).Str("orgResourceId", subOrganization.ResourceID).Msg("Could not set organization limits to unlimited, continuing")
	}

	utils.PrintlnGreen(fmt.Sprintf("+ organization created: %s (depth %d, id %d, parent id %d)", subOrganization.ResourceID, depth, subOrganization.ID, parentId))
	log.Debug().Int64("parent", parentId).Int64("depth", depth).Str("resourceId", subOrganization.ResourceID).Msg("Sub organization created")

	return &orgNode{
		id:         subOrganization.ID,
		resourceID: subOrganization.ResourceID,
		depth:      depth,
	}, nil
}

// distributeDevices creates deviceCount devices, each assigned to a randomly
// picked organization from targets (a slice of organization resource IDs).
func distributeDevices(ctx context.Context, deviceCount int64, targets []string) error {
	if len(targets) == 0 {
		return nil
	}

	for i := int64(0); i < deviceCount; i++ {
		target := targets[rand.Intn(len(targets))]
		name := utils.RandomName()

		_, err := api.CreateDevice(ctx, target, name)
		if err != nil {
			return err
		}

		utils.PrintlnCyan(fmt.Sprintf("+ device created: %s -> org %s (%d/%d)", name, target, i+1, deviceCount))
		log.Debug().Int64("i", i).Str("name", name).Str("orgResourceId", target).Msg("Device created in organization")
	}

	return nil
}
