package knative

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func listAllApps(client clientservingv1.KnServingClient, ctx context.Context) (string, error) {
	appsList, err := client.ListServices(ctx)
	if err != nil {
		zap.S().Errorf("Error while listing apps: %v", err)
		return "", err
	}

	jsonAppList, err := json.Marshal(appsList)
	if err != nil {
		zap.S().Errorf("Error while json marshalling the apps list: %v", err)
		return "", err
	}

	return string(jsonAppList), nil
}

func getAppByName(client clientservingv1.KnServingClient, ctx context.Context, appName string) (string, error) {
	appGetByName, err := client.GetService(ctx, appName)
	if err != nil {
		zap.S().Errorf("Error while listing app: %v", err)
		return "", err
	}
	jsonApp, err := json.Marshal(appGetByName)
	if err != nil {
		zap.S().Errorf("Error while json marshalling the app: %v", err)
		return "", err
	}
	return string(jsonApp), nil
}

func createAppKnative(ctx context.Context, client clientservingv1.KnServingClient, service *servingv1.Service) (err error) {

	err = client.CreateService(ctx, service)
	if err != nil {
		return err
	}
	return nil
}

func deleteApp(client clientservingv1.KnServingClient, ctx context.Context, appName string, timeout time.Duration) error {
	errdelete := client.DeleteService(ctx, appName, timeout)
	if errdelete != nil {
		zap.S().Errorf("Error while deleting app: %v", errdelete)
		return errdelete
	}
	return nil
}
