package knative

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
)

func listAllApps(client clientservingv1.KnServingClient, ctx context.Context) (string, error) {
	appsList, err := client.ListServices(ctx)
	if err != nil {
		zap.S().Errorf("Error while listing apps: %v", err)
		return "", err
	}

	// Encode the apps list in json format
	jsonAppList, err := json.Marshal(appsList)
	if err != nil {
		zap.S().Errorf("Error while json marshalling the apps list: %v", err)
		return "", err
	}

	return string(jsonAppList), nil
}
