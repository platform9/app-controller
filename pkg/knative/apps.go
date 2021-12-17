package knative

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"strconv"

	"knative.dev/client/pkg/kn/commands"
	//	"knative.dev/client/pkg/kn/commands/service"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servinglib "knative.dev/client/pkg/serving"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
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
	jsonAppList, err := json.Marshal(appsList)
	if err != nil {
		log.Error(err, "Error while json marshalling the apps list")
		return "", err
	}

	return string(jsonAppList), nil
}

// Get app by name
func GetAppByName(kubeconfig string, space string, appName string) (apps_list string, err error) {
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

	// Call the knative API to get service by Name
	appGetByName, err := client.GetService(ctx, appName)
	if err != nil {
		log.Error(err, "Error while listing app")
		return "", err
	}

	// Encode the app info in json format
	jsonApp, err := json.Marshal(appGetByName)
	if err != nil {
		log.Error(err, "Error while json marshalling the app")
		return "", err
	}

	return string(jsonApp), nil
}

func containerOfPodSpec(spec *corev1.PodSpec) *corev1.Container {
	if len(spec.Containers) == 0 {
		newContainer := corev1.Container{}
		spec.Containers = append(spec.Containers, newContainer)
	}
	return &spec.Containers[0]
}

func constructService(
	name string,
	namespace string,
	image string,
	env []corev1.EnvVar,
	port string) (service servingv1.Service, err error) {

	service = servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	service.Spec.Template = servingv1.RevisionTemplateSpec{
		Spec: servingv1.RevisionSpec{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				servinglib.UserImageAnnotationKey: "",
			},
		},
	}
	service.Spec.Template.Spec.Containers = []corev1.Container{{}}
	template := &service.Spec.Template
	container := containerOfPodSpec(&template.Spec.PodSpec)
	container.Image = image
	container.Env = env

	if port != "" {
		port_num, err := strconv.Atoi(port)
			if err != nil {
				return service, err
			}
		container.Ports = []corev1.ContainerPort{{
			ContainerPort: int32(port_num),
			Name:          name,
		}}
	}
	return service, nil
}

func serviceExists(ctx context.Context, client clientservingv1.KnServingClient, name string) (bool, error) {
	_, err := client.GetService(ctx, name)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func createAppKnative(
	ctx context.Context,
	client clientservingv1.KnServingClient,
	service *servingv1.Service) (err error) {

	err = client.CreateService(ctx, service)
	if err != nil {
		return err
	}
	return nil
}

func CreateApp(
	kubeconfig string,
	appName string,
	space string,
	image string,
	env []corev1.EnvVar,
	port string) (err error) {

	// Initialize the knative parameters
	knParams := &commands.KnParams{}
	knParams.KubeCfgPath = kubeconfig
	knParams.Initialize()

	// Fetch the knative serving client for a given knative space
	client, err := knParams.NewServingClient(space)
	if err != nil {
		log.Error(err, "Error while creating a knative serving client")
		return err
	}

	// Create an empty context, required for knative APIs
	ctx := context.Background()

	service, err := constructService(appName, space, image, env, port)
	if err != nil {
		log.Error(err, "Error while creating the service object")
		return err
	}

	fmt.Printf("Service : %v\n", service)

	serviceExists, err := serviceExists(ctx, client, service.Name)
	if err != nil {
		log.Error(err, "Error while checking for service existence")
		return err
	}

	if serviceExists {
		return fmt.Errorf("Service already exists")
	} else {
		err = createAppKnative(ctx, client, &service)
	}
	if err != nil {
		return err
	}

	return nil
}

// Delete an app by name
func DeleteApp(kubeconfig string, space string, appName string) error {
	// Initialize the knative parameters
	knParams := &commands.KnParams{}
	knParams.KubeCfgPath = kubeconfig
	knParams.Initialize()

	// Fetch the knative serving client for a given knative space
	client, err := knParams.NewServingClient(space)
	if err != nil {
		log.Error(err, "Error while creating a knative serving client")
		return err
	}

	// Create an empty context, required for knative APIs
	ctx := context.Background()
	/* To delete service without any wait.
	timeout -- duration to wait for a delete operation to finish.
	*/
	var timeout = time.Duration(0)

	// Call the knative API to delete service by Name
	errdelete := client.DeleteService(ctx, appName, timeout)
	if errdelete != nil {
		log.Error(errdelete, "Error while deleting app")
		return err
	}

	return nil
}
