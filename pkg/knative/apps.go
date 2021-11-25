package knative

import (
	"context"
	"encoding/json"

	"knative.dev/client/pkg/kn/commands"
	log "github.com/sirupsen/logrus"
)

func GetApps(kubeconfig string, space string) (apps_list string, err error) {
	// Initialize the knative parameters
	knParams := &commands.KnParams{}
	knParams.KubeCfgPath = kubeconfig
	knParams.Initialize()

	// Fetch the knative serving client for a given knative space
	client, err := knParams.NewServingClient(space)
	if err != nil {
		log.Error(err, "Error while creating a knative serving client")
		return "", err
	}

	// Create an empty context, required for knative APIs
	ctx := context.Background()

	// Call the knative API
	appsList, err := client.ListServices(ctx)
	if err != nil {
		log.Error(err, "Error while listing apps")
		return "", err
	}


	// Encode the apps list in json format
	jsonAppList, err:= json.Marshal(appsList)
        if err != nil {
                log.Error(err, "Error while json marshalling the apps list")
                return "", err
        }


	return string(jsonAppList), nil
}
