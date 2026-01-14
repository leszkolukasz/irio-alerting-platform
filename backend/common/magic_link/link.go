package magic_link

import (
	"alerting-platform/common/config"
	"fmt"
)

const ResolveEndpointPath = "/api/v1/incidents/resolve"

func GenerateResolveLink(incidentID string, serviceID uint64, email string, secretKey []byte, apiHost string, apiPort int) (string, error) {
	token, err := GenerateToken(incidentID, serviceID, email, secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate token for link: %w", err)
	}

	var link string
	if config.GetConfig().Env != config.PROD {
		link = fmt.Sprintf("%s:%d%s/%s", apiHost, apiPort, ResolveEndpointPath, token)
	} else {
		link = fmt.Sprintf("%s%s/%s", apiHost, ResolveEndpointPath, token)
	}

	return link, nil
}
