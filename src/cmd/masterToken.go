package cmd

import (
	"context"
	"corvina/corvina-seed/src/seed/keycloak"
	"fmt"
)

func MasterToken(ctx context.Context) error {
	// Get the master token
	masterToken, err := keycloak.MasterToken(ctx)
	if err != nil {
		return err
	}

	// Print the master token
	fmt.Println(*masterToken)

	return nil
}
