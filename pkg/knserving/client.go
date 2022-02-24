package knserving

import (
	"go.uber.org/zap"
	"knative.dev/client/pkg/kn/commands"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
)

func GetKnativeServingClient(kubeconfigPath, namespace string) (clientservingv1.KnServingClient, error) {
	// Initialize the knative parameters
	knParams := &commands.KnParams{}
	knParams.KubeCfgPath = kubeconfigPath
	knParams.Initialize()

	// Fetch the knative serving client for a given knative space
	client, err := knParams.NewServingClient(namespace)
	if err != nil {
		zap.S().Errorf("Error while creating a knative serving client: %v", err)
		return nil, err
	}
	return client, nil
}
