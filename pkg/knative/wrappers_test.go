package knative

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/assert"
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
