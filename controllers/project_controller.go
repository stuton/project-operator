/*
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
*/

package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	operationsv1alpha1 "github.com/stuton/project-operator/api/v1alpha1"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operations.operator.io,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operations.operator.io,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operations.operator.io,resources=projects/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=namespaces;resourcequotas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,resourceNames=edit,verbs=bind

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info(fmt.Sprintf("Starting reconcile loop for %v", req.NamespacedName))
	defer log.Info(fmt.Sprintf("Finished reconcile loop for %v", req.NamespacedName))

	project := operationsv1alpha1.Project{}
	if err := r.Client.Get(ctx, req.NamespacedName, &project); err != nil {
		log.Error(err, "unable to fetch Project")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	for _, environment := range project.Spec.Environments {
		result, err := r.ensureNamespace(ctx, req, project, r.describeNamespace(project, environment), log)
		if result != nil {
			return *result, err
		}

		result, err = r.ensureResourceQuota(ctx, req, project, r.describeResourceQuota(project, environment), log)
		if result != nil {
			return *result, err
		}

		for _, member := range project.Spec.Members {
			result, err := r.ensureRoleBinding(ctx, req, project, r.describeRoleBinding(project, environment, member), log)
			if result != nil {
				return *result, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operationsv1alpha1.Project{}).
		Complete(r)
}

func (r *ProjectReconciler) ensureNamespace(ctx context.Context, req ctrl.Request, project operationsv1alpha1.Project, n *corev1.Namespace, log logr.Logger) (*reconcile.Result, error) {
	namespace := &corev1.Namespace{}

	err := r.Client.Get(ctx, types.NamespacedName{
		Name: project.Name + "-" + n.Name,
	}, namespace)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Namespace", "Name", n.Name)

		if err = r.Client.Create(ctx, n); err != nil {
			log.Error(err, "Failed to create new Namespace", "Name", n.Name)
			return &reconcile.Result{}, err
		}

		return nil, nil
	}

	if err != nil {
		log.Error(err, "Failed to get Namespace")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ProjectReconciler) describeNamespace(project operationsv1alpha1.Project, env operationsv1alpha1.Environment) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   project.Name + "-" + env.Name,
			Labels: projectlabels(project, "namespace"),
		},
	}
}

func (r *ProjectReconciler) ensureResourceQuota(ctx context.Context, req ctrl.Request, project operationsv1alpha1.Project, rq *corev1.ResourceQuota, log logr.Logger) (*reconcile.Result, error) {
	resourceQuota := &corev1.ResourceQuota{}

	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      rq.Name,
		Namespace: project.Name + "-" + rq.Name,
	}, resourceQuota)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Resourse Quota", "Name", rq.Name)

		if err = r.Client.Create(ctx, rq); err != nil {
			log.Error(err, "Failed to create new Resourse Quota", "Name", rq.Name)
			return &reconcile.Result{}, err
		}

		return nil, nil
	}

	if err != nil {
		log.Error(err, "Failed to get Namespace")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ProjectReconciler) describeResourceQuota(project operationsv1alpha1.Project, env operationsv1alpha1.Environment) *corev1.ResourceQuota {
	return &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: project.Name + "-" + env.Name,
			Name:      "resource-quota",
			Labels:    projectlabels(project, "resource-quota"),
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				"requests.cpu":    env.Resources.Requests.CPU,
				"requests.memory": env.Resources.Requests.Memory,
				"limits.cpu":      env.Resources.Limits.CPU,
				"limits.memory":   env.Resources.Limits.Memory,
			},
		},
	}
}

func (r *ProjectReconciler) ensureRoleBinding(ctx context.Context, req ctrl.Request, project operationsv1alpha1.Project, rb *rbacv1.RoleBinding, log logr.Logger) (*reconcile.Result, error) {
	roleBinding := &rbacv1.RoleBinding{}

	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      rb.Name,
		Namespace: project.Name + "-" + rb.Name,
	}, roleBinding)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Role Binding", "Name", rb.Name)

		if err = r.Client.Create(ctx, rb); err != nil {
			log.Error(err, "Failed to create new Role Binding", "Name", rb.Name)
			return &reconcile.Result{}, err
		}
		// Creation was successful
		return nil, nil
	}

	if err != nil {
		log.Error(err, "Failed to get Namespace")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *ProjectReconciler) describeRoleBinding(project operationsv1alpha1.Project, env operationsv1alpha1.Environment, member string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: project.Name + "-" + env.Name,
			Name:      member,
			Labels:    projectlabels(project, "role-binding"),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "edit",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      member,
				Namespace: "users",
			},
		},
	}
}

func projectlabels(project operationsv1alpha1.Project, component string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       project.Name,
		"app.kubernetes.io/component":  component,
		"app.kubernetes.io/managed-by": "project-operator",
	}
}
