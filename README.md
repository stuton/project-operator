# project-operator
The concept of the project is taken from OpenShift [projects](https://docs.openshift.com/online/pro/architecture/core_concepts/projects_and_users.html#projects).
Because in kubernetes there is no such built-in object called a project. It is allowed to create an entity over namespaces such as Resource quota, Limit Range and Role Bindings

## Description
In general, creating a new namespace for a new application or command is not a trivial task, as it may seem. And in fact, in order for us to create a namespace for a new application, we need:
* create a resource quota that will limit resources
* number of pods that can be run in namespace
* you need to create rights for developers (for example, in production, just to see if on a dev or test the rights to interact with these objects)


## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Deploy the controller to the cluster:

```sh
make deploy IMG=winshiftq/project-operator:v0.0.1
```
2. Install Project samples:

```sh
kubectl apply -f config/samples/
```

### Undeploy controller
UnDeploy the controller from the cluster:

```sh
make undeploy
```

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.


```
k get ns
...
project-operator-system   Active   67m
project-sample-dev        Active   66m
project-sample-prod       Active   66m
project-sample-test       Active   66m
```

Now we have resource quotas objects in each new namespaces:

```
k get resourcequotas -A

project-sample-dev    resource-quota   67m   requests.cpu: 0/1, requests.memory: 0/1Gi   limits.cpu: 0/1, limits.memory: 0/1Gi
project-sample-prod   resource-quota   67m   requests.cpu: 0/2, requests.memory: 0/4Gi   limits.cpu: 0/2, limits.memory: 0/4Gi
project-sample-test   resource-quota   67m   requests.cpu: 0/1, requests.memory: 0/1Gi   limits.cpu: 0/1, limits.memory: 0/1Gi
```
We also have rights for each member in these namespaces:
```
k get rolebindings.rbac.authorization.k8s.io -A

project-operator-system   project-operator-leader-election-rolebinding        Role/project-operator-leader-election-role            69m
project-sample-dev        bray.ferguson                                       ClusterRole/edit                                      68m
project-sample-dev        clarke.armstrong                                    ClusterRole/edit                                      68m
project-sample-dev        hansen.rivera                                       ClusterRole/edit                                      68m
project-sample-prod       bray.ferguson                                       ClusterRole/edit                                      68m
project-sample-prod       clarke.armstrong                                    ClusterRole/edit                                      68m
project-sample-prod       hansen.rivera                                       ClusterRole/edit                                      68m
project-sample-test       bray.ferguson                                       ClusterRole/edit                                      68m
project-sample-test       clarke.armstrong                                    ClusterRole/edit                                      68m
project-sample-test       hansen.rivera                                       ClusterRole/edit                                      68m
```


## Code structure
```
├── api
│   └── v1alpha1
│       ├── groupversion_info.go
│       ├── project_types.go
│       └── zz_generated.deepcopy.go
├── bin
│   ├── controller-gen
│   ├── k8s
│   │   └── 1.26.0-darwin-arm64
│   │       ├── etcd
│   │       ├── kube-apiserver
│   │       └── kubectl
│   ├── kustomize
│   └── setup-envtest
├── config
│   ├── crd
│   │   ├── bases
│   │   │   └── operations.operator.io_projects.yaml
│   │   ├── kustomization.yaml
│   │   ├── kustomizeconfig.yaml
│   │   └── patches
│   │       ├── cainjection_in_projects.yaml
│   │       └── webhook_in_projects.yaml
│   ├── default
│   │   ├── kustomization.yaml
│   │   ├── manager_auth_proxy_patch.yaml
│   │   └── manager_config_patch.yaml
│   ├── manager
│   │   ├── kustomization.yaml
│   │   └── manager.yaml
│   ├── manifests
│   │   └── kustomization.yaml
│   ├── prometheus
│   │   ├── kustomization.yaml
│   │   └── monitor.yaml
│   ├── rbac
│   │   ├── auth_proxy_client_clusterrole.yaml
│   │   ├── auth_proxy_role.yaml
│   │   ├── auth_proxy_role_binding.yaml
│   │   ├── auth_proxy_service.yaml
│   │   ├── kustomization.yaml
│   │   ├── leader_election_role.yaml
│   │   ├── leader_election_role_binding.yaml
│   │   ├── project_editor_role.yaml
│   │   ├── project_viewer_role.yaml
│   │   ├── role.yaml
│   │   ├── role_binding.yaml
│   │   └── service_account.yaml
│   ├── samples
│   │   ├── kustomization.yaml
│   │   └── operations_v1alpha1_project.yaml
│   └── scorecard
│       ├── bases
│       │   └── config.yaml
│       ├── kustomization.yaml
│       └── patches
│           ├── basic.config.yaml
│           └── olm.config.yaml
├── controllers
│   ├── project_controller.go
│   └── suite_test.go
├── cover.out
├── go.mod
├── go.sum
├── Dockerfile
├── Makefile
├── README.md
└── main.go
```

## Repository overview

* [api/](./api):	Contains the api definition
* [bin/](./bin): This directory contains useful binaries such as the manager which is used to run your project locally and the kustomize utility used for the project configuration. For other language types, it might have other binaries useful for developing your operator.
* [config/](./config):	Contains configuration files to launch your project on a cluster. Plugins might use it to provide functionality. For example, for the CLI to help create your operator bundle it will look for the CRD’s and CR’s which are scaffolded in this directory. You will also find all Kustomize YAML definitions as well.
* [config/crd/](./config/crd):	Contains the Custom Resources Definitions.
* [config/default/](./config/default):	Contains a Kustomize base for launching the controller in a standard configuration.
* [config/manager/](./config/manager):	Contains the manifests to launch your operator project as pods on the cluster.
* [config/manifests/](./config/manifests):	Contains the base to generate your OLM manifests in the bundle directory.
* [config/prometheus/](./config/prometheus):	Contains the manifests required to enable project to serve metrics to Prometheus such as the ServiceMonitor resource.
* [config/scorecard/](./config/scorecard):	Contains the manifests required to allow you test your project with Scorecard.
* [config/rbac/](./config/rbac):	Contains the RBAC permissions required to run your project.
* [config/samples/](./config/samples):	Contains the Custom Resources.
* [controllers](./controllers):	Contains the controllers.
* main.go:	Implements the project initialization.

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

