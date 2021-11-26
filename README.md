# fast-path

Service for interacting with knative

## TL;DR

### Build and run service

fast-path can be run as a binary on linux machine.

```
$ make build
$ ./bin/fast-path
```

The kubeconfig of the cluster on which the knative is installed, should be placed in
```
/etc/pf9/fast-path/config.yaml
```

Example contents of this file look as follows:

```
# cat /etc/pf9/fast-path/config.yaml
kubeconfig:
  file: "/home/ubuntu/kubeconfig/aws3.yaml"
```


### API test

#### GET all apps in default namespace

```
# curl http://127.0.0.1:6112/v1/apps/default | jq .
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  7466    0  7466    0     0   8289      0 --:--:-- --:--:-- --:--:--  8286
{
  "kind": "ServiceList",
  "apiVersion": "serving.knative.dev/v1",
  "metadata": {
    "resourceVersion": "12353051"
  },
  "items": [
    {
      "kind": "Service",
      "apiVersion": "serving.knative.dev/v1",
      "metadata": {
        "name": "hello",
        "namespace": "default",
        "uid": "3686f908-14de-456c-b4d0-83ee7c825afd",
        "resourceVersion": "33948",
        "generation": 1,
        "creationTimestamp": "2021-11-15T08:12:23Z",
        "annotations": {
          "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"serving.knative.dev/v1\",\"kind\":\"Service\",\"metadata\":{\"annotations\":{},\"name\":\"hello\",\"namespace\":\"default\"},\"spec\":{\"template\":{\"metadata\":{\"name\":\"hello-world\"},\"spec\":{\"containers\":[{\"env\":[{\"name\":\"TARGET\",\"value\":\"World\"}],\"image\":\"gcr.io/knative-samples/helloworld-go\",\"ports\":[{\"containerPort\":8080}]}]}}}}\n",
          "serving.knative.dev/creator": "anup+751@platform9.com",
          "serving.knative.dev/lastModifier": "anup+751@platform9.com"
        },
        "managedFields": [
          {
            "manager": "kubectl",
            "operation": "Update",
            "apiVersion": "serving.knative.dev/v1",
            "time": "2021-11-15T08:12:23Z",
            "fieldsType": "FieldsV1",
            "fieldsV1": {
              "f:metadata": {
                "f:annotations": {
                  ".": {},
                  "f:kubectl.kubernetes.io/last-applied-configuration": {}
                }
              },
              "f:spec": {
                ".": {},
                "f:template": {
                  ".": {},
                  "f:metadata": {
                    ".": {},
                    "f:name": {}
                  },
                  "f:spec": {
                    ".": {},
                    "f:containers": {}
                  }
                }
              }
            }
          },
          {
            "manager": "controller",
            "operation": "Update",
            "apiVersion": "serving.knative.dev/v1",
            "time": "2021-11-15T08:12:45Z",
            "fieldsType": "FieldsV1",
            "fieldsV1": {
              "f:status": {
                ".": {},
                "f:address": {
                  ".": {},
                  "f:url": {}
                },
                "f:conditions": {},
                "f:latestCreatedRevisionName": {},
                "f:latestReadyRevisionName": {},
                "f:observedGeneration": {},
                "f:traffic": {},
                "f:url": {}
              }
            }
          }
        ]
      },
      "spec": {
        "template": {
          "metadata": {
            "name": "hello-world",
            "creationTimestamp": null
          },
          "spec": {
            "containers": [
              {
                "name": "user-container",
                "image": "gcr.io/knative-samples/helloworld-go",
                "ports": [
                  {
                    "containerPort": 8080,
                    "protocol": "TCP"
                  }
                ],
                "env": [
                  {
                    "name": "TARGET",
                    "value": "World"
                  }
                ],
                "resources": {},
                "readinessProbe": {
                  "tcpSocket": {
                    "port": 0
                  },
                  "successThreshold": 1
                }
              }
            ],
            "enableServiceLinks": false,
            "containerConcurrency": 0,
            "timeoutSeconds": 300
          }
        },
        "traffic": [
          {
            "latestRevision": true,
            "percent": 100
          }
        ]
      },
      "status": {
        "observedGeneration": 1,
        "conditions": [
          {
            "type": "ConfigurationsReady",
            "status": "True",
            "lastTransitionTime": "2021-11-15T08:12:45Z"
          },
          {
            "type": "Ready",
            "status": "True",
            "lastTransitionTime": "2021-11-15T08:12:45Z"
          },
          {
            "type": "RoutesReady",
            "status": "True",
            "lastTransitionTime": "2021-11-15T08:12:45Z"
          }
        ],
        "latestReadyRevisionName": "hello-world",
        "latestCreatedRevisionName": "hello-world",
        "url": "http://hello.default.52.8.39.229.sslip.io",
        "address": {
          "url": "http://hello.default.svc.cluster.local"
        },
        "traffic": [
          {
            "revisionName": "hello-world",
            "latestRevision": true,
            "percent": 100
          }
        ]
      }
    },
  ]
}
```


#### POST (create) an app in default namespace

```
# curl -X POST http://127.0.0.1:6112/v1/apps --data-binary '{"name":"test-hello-world", "space":"default", "image":"gcr.io/knative-samples/helloworld-go"}' -H "Content-Type:application/json"
```
where,
name = name of the app
space = namespace in which app has to be created
image = container image of the app
