package knative

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	v1 "knative.dev/client/pkg/serving/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingv1fake "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1/fake"
)

var testNamespace = "test"

func setup() (serving servingv1fake.FakeServingV1, client v1.KnServingClient) {
	serving = servingv1fake.FakeServingV1{Fake: &clienttesting.Fake{}}
	client = v1.NewKnServingClient(&serving, testNamespace)
	return
}

func newService(name string) *servingv1.Service {
	return &servingv1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace}}
}

func TestListAllApps(t *testing.T) {
	serving, client := setup()

	t.Run("list all apps when multiple apps are present", func(t *testing.T) {
		service1 := newService("service-1")
		service2 := newService("service-2")
		service3 := newService("service-3")
		serving.AddReactor("list", "services",
			func(a clienttesting.Action) (bool, runtime.Object, error) {
				assert.Equal(t, testNamespace, a.GetNamespace())
				return true, &servingv1.ServiceList{Items: []servingv1.Service{*service1, *service2, *service3}}, nil
			})
		allApps, err := listAllApps(client, context.Background())
		assert.NilError(t, err)
		var appInfo servingv1.ServiceList
		unmarshalErr := json.Unmarshal([]byte(allApps), &appInfo)
		if unmarshalErr != nil {
			t.Errorf("error parsing json response\n%s\n", unmarshalErr.Error())
		}
		if len(appInfo.Items) != 3 {
			t.Errorf("expected %d items got %d", 3, len(appInfo.Items))
		}
	})
}

func TestListAppsEmpty(t *testing.T) {
	serving, client := setup()
	t.Run("list all apps when no apps are present", func(t *testing.T) {
		serving.AddReactor("list", "services",
			func(a clienttesting.Action) (bool, runtime.Object, error) {
				assert.Equal(t, testNamespace, a.GetNamespace())
				return true, &servingv1.ServiceList{Items: []servingv1.Service{}}, nil
			})
		allApps, err := listAllApps(client, context.Background())
		assert.NilError(t, err)
		var appInfo servingv1.ServiceList
		unmarshalErr := json.Unmarshal([]byte(allApps), &appInfo)
		if unmarshalErr != nil {
			t.Errorf("error parsing json response\n%s\n", unmarshalErr.Error())
		}
		if len(appInfo.Items) != 0 {
			t.Errorf("expected 0 length, got non zero: %d\n", len(appInfo.Items))
		}
	})
}

func TestGetAppByName(t *testing.T) {
	serving, client := setup()
	appName := "test-service-that-exists"
	nonExistentApp := "service-that-doesnot-exist"
	serving.AddReactor("get", "services", func(action clienttesting.Action) (bool, runtime.Object, error) {
		service := newService(appName)
		name := action.(clienttesting.GetAction).GetName()

		assert.Assert(t, name != "")
		assert.Equal(t, testNamespace, action.GetNamespace())
		if name == appName {
			return true, service, nil
		}
		return true, nil, errors.NewNotFound(servingv1.Resource("service"), name)
	})
	t.Run("get a service that is present", func(t *testing.T) {
		app, err := getAppByName(client, context.Background(), appName)
		assert.NilError(t, err, nil)
		var appDetails servingv1.Service
		if unMarErr := json.Unmarshal([]byte(app), &appDetails); unMarErr != nil {
			t.Errorf("error unmarshaling %s\n", unMarErr.Error())
		}
		assert.Equal(t, appName, appDetails.Name, "service name should be equal")
	})

	t.Run("get a service that does not exist", func(t *testing.T) {
		app, err := getAppByName(client, context.Background(), nonExistentApp)
		assert.Assert(t, app == "", "no service should be returned")
		assert.ErrorContains(t, err, "not found")
		assert.ErrorContains(t, err, nonExistentApp)
	})
}

func TestCreateApp(t *testing.T) {
	serving, client := setup()
	newApp := "new-app"
	unknownApp := "unknown-app"
	newDeployedApp := newService(newApp)
	unknownService := newService(unknownApp)
	serving.AddReactor("create", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(clienttesting.CreateAction).GetObject().(metav1.Object).GetName()
			if name == newDeployedApp.Name {
				newDeployedApp.Generation = 2
				return true, newDeployedApp, nil
			}
			return true, nil, fmt.Errorf("error while creating service %s", name)
		})
	t.Run("reate a service without errors", func(t *testing.T) {
		err := createAppKnative(context.Background(), client, newDeployedApp)
		assert.NilError(t, err)
		assert.Equal(t, newDeployedApp.Generation, int64(2))
		assert.Equal(t, newDeployedApp.Name, newApp)
	})
	t.Run("simulate an error", func(t *testing.T) {
		err := createAppKnative(context.Background(), client, unknownService)
		assert.ErrorContains(t, err, "unknown")
	})
}

func TestDeleteApp(t *testing.T) {
	serving, client := setup()
	const (
		serviceName            = "test-service"
		nonExistingServiceName = "no-service"
	)
	serving.AddReactor("get", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.GetAction).GetName()
			if name == serviceName {
				// Don't handle existing service, just continue to next
				return false, nil, nil
			}
			return true, nil, errors.NewNotFound(servingv1.Resource("service"), name)
		})

	serving.AddReactor("delete", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.DeleteAction).GetName()

			assert.Assert(t, name != "")
			assert.Equal(t, testNamespace, a.GetNamespace())
			if name == serviceName {
				return true, nil, nil
			}
			return false, nil, nil
		})
	t.Run("delete existing service returns no error", func(t *testing.T) {
		err := deleteApp(client, context.Background(), serviceName, 0)
		assert.NilError(t, err)
	})
	t.Run("trying to delete non-existing service returns error", func(t *testing.T) {
		err := client.DeleteService(context.Background(), nonExistingServiceName, 0)
		println(err.Error())
		assert.ErrorContains(t, err, "not found")
		assert.ErrorContains(t, err, nonExistingServiceName)
	})
}
