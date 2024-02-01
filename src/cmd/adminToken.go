package cmd

import (
	"context"
	"corvina/corvina-seed/src/seed/keycloak"
	"fmt"
)

func AdminToken(ctx context.Context) error {
	// Get the admin token
	adminToken, err := keycloak.AdminToken(ctx)
	if err != nil {
		return err
	}

	// Print the admin token
	fmt.Println(*adminToken)

	return nil
}
