/*
 * project.go
 */

package authorizer

import (
	"context"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

func GetProjectID() (string, error) {
	ctx := context.Background()

	credentials, err := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if err != nil {
		return "", err
	}

	return credentials.ProjectID, nil
}
