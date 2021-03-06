package knative

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/platform9/app-controller/pkg/options"
	"github.com/platform9/app-controller/pkg/util"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/cmd/create"
	"knative.dev/client/pkg/kn/commands"
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
		zap.S().Errorf("Error while creating a knative serving client: %v", err)
		return "", err
	}

	// Create an empty context, required for knative APIs
	ctx := context.Background()

	// Call the knative API wrapper
	return listAllApps(client, ctx)
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
		zap.S().Errorf("Error while creating a knative serving client: %v", err)
		return "", err
	}

	// Create an empty context, required for knative APIs
	ctx := context.Background()

	// Call the knative API wrapper to get service by Name
	return getAppByName(client, ctx, appName)

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
	port string,
	secretname string) (service servingv1.Service, err error) {

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

	if secretname != "" {
		template.Spec.PodSpec.ImagePullSecrets = []corev1.LocalObjectReference{{
			Name: secretname,
		}}
	}

	if port != "" {
		port_num, err := strconv.Atoi(port)
		if err != nil {
			return service, err
		}
		container.Ports = []corev1.ContainerPort{{
			ContainerPort: int32(port_num),
			Name:          "",
		}}
	}

	max_scale := options.GetConstraintMaxScale()
	servinglib.UpdateMaxScale(template, max_scale)
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

// handleDockerCfgJSONContent serializes a ~/.docker/config.json file
func handleDockerCfgJSONContent(username, password, server string) ([]byte, error) {
	dockerConfigAuth := create.DockerConfigEntry{
		Username: username,
		Password: password,
		Auth:     encodeDockerConfigFieldAuth(username, password),
	}
	dockerConfigJSON := create.DockerConfigJSON{
		Auths: map[string]create.DockerConfigEntry{server: dockerConfigAuth},
	}

	return json.Marshal(dockerConfigJSON)
}

// encodeDockerConfigFieldAuth returns base64 encoding of the username and password string
func encodeDockerConfigFieldAuth(username, password string) string {
	fieldValue := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(fieldValue))
}

// Constructor for corev1.Secret object
func newSecretObj(name, namespace string, secretType corev1.SecretType) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: secretType,
		Data: map[string][]byte{},
	}
}

func createDockerRegistry(so create.CreateSecretDockerRegistryOptions) (*corev1.Secret, error) {
	secretDockerRegistry := newSecretObj(so.Name, so.Namespace, corev1.SecretTypeDockerConfigJson)
	dockerConfigJSONContent, err := handleDockerCfgJSONContent(so.Username, so.Password, so.Server)
	if err != nil {
		return nil, err
	}
	secretDockerRegistry.Data[corev1.DockerConfigJsonKey] = dockerConfigJSONContent

	return secretDockerRegistry, nil
}

func extractRegistryURL(url string) (containerURL string, err error) {

	url = strings.TrimPrefix(url, util.HTTPURL)
	url = strings.TrimPrefix(url, util.HTTPSURL)

	urlList := strings.Split(url, "/")

	if urlList[0] == util.DockerURL {
		return util.DockerServerURL, nil
	}
	if strings.Contains(urlList[0], util.AWSURL) {
		return util.HTTPSURL + urlList[0], nil
	}
	if strings.Contains(urlList[0], util.GCRURL) {
		return util.HTTPSURL + urlList[0], nil
	}
	if strings.Contains(urlList[0], util.ACRURL) {
		return util.HTTPSURL + urlList[0], nil
	}

	return "", fmt.Errorf("Incorrect image format")
}

// Inject container secrets into the namespace
func injectContainerImageSecrets(
	kubeconfig string,
	space string,
	secretname string,
	username string,
	password string,
	image string) (err error) {

	server, err := extractRegistryURL(image)
	if err != nil {
		zap.S().Errorf("Error while extracting registry URL from the image URL: %v", err)
		return err
	}

	// create config structure instance from the kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		zap.S().Errorf("Error while creating config object from kubeconfig: %v", err)
		return err
	}

	// create clientset from the kubeconfig in-mem structure
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		zap.S().Errorf("Error while creating clientset: %v", err)
		return err
	}

	// Create in-mem secretDockerRegistryOptions that are used to create the secretDockerRegistry structure
	var secretDockerRegistry *corev1.Secret = nil
	secretDockerRegistryOptions := create.CreateSecretDockerRegistryOptions{
		Name:       secretname,
		Username:   username,
		Password:   password,
		Server:     server,
		AppendHash: false,
		Namespace:  space,
	}

	// Validate the secretDockerRegistryOptions
	err = secretDockerRegistryOptions.Validate()
	if err != nil {
		zap.S().Errorf("Error Validating secret options: %v", err)
		return err
	}

	// Create in-mem secretDockerRegistry using options.
	secretDockerRegistry, err = createDockerRegistry(secretDockerRegistryOptions)

	createOptions := metav1.CreateOptions{}
	// Fire secret creation CoreV1 API.
	secretDockerRegistry, err = clientset.CoreV1().Secrets(space).Create(context.TODO(), secretDockerRegistry, createOptions)
	if err != nil {
		zap.S().Errorf("Error creating docker secrets: %v", err)
		return err
	}

	return nil
}

func CreateApp(
	kubeconfig string,
	appname string,
	space string,
	image string,
	env []corev1.EnvVar,
	port string,
	secretname string,
	username string,
	password string) (err error) {

	// Initialize the knative parameters
	knParams := &commands.KnParams{}
	knParams.KubeCfgPath = kubeconfig
	knParams.Initialize()

	// Fetch the knative serving client for a given knative space
	client, err := knParams.NewServingClient(space)
	if err != nil {
		zap.S().Errorf("Error while creating a knative serving client: %v", err)
		return err
	}

	// Create an empty context, required for knative APIs
	ctx := context.Background()

	// Check for maximum apps deploy limit.
	stopDeploy, err := maxAppDeployed(kubeconfig, space)
	if err != nil {
		zap.S().Errorf("Error while checking maximum app deployed: %v", err)
		return err
	}

	if stopDeploy {
		zap.S().Errorf("Maximum Apps deploy limit reached!!")
		return fmt.Errorf("Maximum App deploy limit reached!")
	}

	// If container secret info exists, create a secret in the k8s cluster.
	if (username != "") &&
		(password != "") {
		err = injectContainerImageSecrets(kubeconfig, space, secretname, username, password, image)
		if err != nil {
			zap.S().Errorf("Error while injecting the secrets object: %v", err)
		}
	} else {
		// Secret name has no value where username and password don't exist.
		secretname = ""
	}

	service, err := constructService(appname, space, image, env, port, secretname)
	if err != nil {
		zap.S().Errorf("Error while creating the service object: %v", err)
		return err
	}

	zap.S().Debugf("Service : %v\n", service)

	serviceExists, err := serviceExists(ctx, client, service.Name)
	if err != nil {
		zap.S().Errorf("Error while checking for service existence: %v", err)
		return err
	}

	if serviceExists {
		zap.S().Error("Service already exists.")
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
		zap.S().Errorf("Error while creating a knative serving client: %v", err)
		return err
	}

	// Create an empty context, required for knative APIs
	ctx := context.Background()
	/* To delete service without any wait.
	timeout -- duration to wait for a delete operation to finish.
	*/
	var timeout = time.Duration(0)

	// Call the knative API wrapper to delete service by Name
	err = deleteApp(client, ctx, appName, timeout)
	if err != nil {
                zap.S().Errorf("Error while deleting the app: %v", err)
                return err
        }

	err = deleteSecret(kubeconfig, space, appName)
	if err != nil {
		//Not returning error as it's not needed to show this to the user
                zap.S().Debugf("Error while deleting the app secret: %v", err)
        }

	return nil
}

//Delete the secret associated with the app being deleted
//in case the app is deployed from private registry
func deleteSecret(kubeconfig string, space string, appName string) error {
	// create config structure instance from the kubeconfig
        config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
                zap.S().Debugf("Error creating config object from kubeconfig: %v ", err)
                return err
        }

        // create clientset from the kubeconfig in-memory structure
        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
                zap.S().Debugf("Error creating clientset: %v ", err)
                return err
        }

        deleteOptions := metav1.DeleteOptions{}
        // Fire secret deletion CoreV1 API.
	//Secret name is same as the app name.
	//This will fail for public registry as secret won't be present.
	//But this can be treated as no-op.
        err = clientset.CoreV1().Secrets(space).Delete(context.TODO(), appName, deleteOptions)
        if err != nil {
		zap.S().Debugf("Error deleting the secret: %v", err)
                return err
        }

	return nil
}

// Check if the apps deployed exceeds maxAppDeployCount.
func maxAppDeployed(kubeconfig string, space string) (bool, error) {
	get_apps, errMax := GetApps(kubeconfig, space)
	if errMax != nil {
		zap.S().Errorf("Error while listing apps: %v", errMax)
		return false, errMax
	}

	var appList map[string]interface{}

	err := json.Unmarshal([]byte(get_apps), &appList)
	if err != nil {
		zap.S().Errorf("Failed to Unmarshal: %v", err)
		return false, err
	}

	max_app := options.GetConstraintMaxAppDeploy()

	if len(appList["items"].([]interface{})) >= max_app {
		return true, nil
	}
	return false, nil
}
