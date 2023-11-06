package gh

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// GenerateAppToken creates an access token for the given GitHub app in the given GitHub org. The app must be installed
// in the org. The returned access token is valid for one hour.
// (for details see: https://docs.github.com/en/rest/apps/apps#create-an-installation-access-token-for-an-app)
func GenerateAppToken(org string, appId string, appKey string) (string, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(appKey))
	if err != nil {
		return "", fmt.Errorf("couldn't parse key: %w", err)
	}

	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		Issuer:    appId,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Minute)), // We only need it to generate an access token
	})

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("couldn't sign token: %w", err)
	}

	ctx := context.Background()
	github := NewGitHubClient(tokenString)

	installation, _, err := github.client.Apps.FindOrganizationInstallation(ctx, org)
	if err != nil {
		return "", fmt.Errorf("couldn't find app installation for org: %w", err)
	}

	accessToken, _, err := github.client.Apps.CreateInstallationToken(ctx, *installation.ID)
	if err != nil {
		return "", fmt.Errorf("couldn't create access token: %w", err)
	}

	return accessToken.GetToken(), nil
}
