/*
Copyright 2024.

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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	homelabv1alpha1 "gitlab.srv-lab.server.home/homelab/iac/operators/rq-operator/api/v1alpha1"
)

// ResourceQuotaConfigReconciler reconciles a ResourceQuotaConfig object
type ResourceQuotaConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=homelab.server.home,resources=resourcequotaconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=homelab.server.home,resources=resourcequotaconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=homelab.server.home,resources=resourcequotaconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=resourcequotas,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ResourceQuotaConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ResourceQuotaConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling ResourceQuotaConfig")

	// Fetch the ResourceQuotaConfig instance
	instance := &homelabv1alpha1.ResourceQuotaConfig{}
	if err := r.fetchResourceQuotaConfig(ctx, req, instance); err != nil {
		log.Error(err, "Failed to fetch ResourceQuotaConfig")

		return ctrl.Result{}, err
	}

	// Fetch namespaces that match the selector or if not set all namespaces
	namespaces := &corev1.NamespaceList{}
	if err := r.fetchNamespaces(ctx, instance.Spec.NamespaceSelector, namespaces); err != nil {
		log.Error(err, "Failed to fetch namespaces")

		return ctrl.Result{}, nil
	}

	// Now that we have the namespaces, we can create or delete ResourceQuotas
	if err := r.processNamespaces(ctx, namespaces, instance); err != nil {
		log.Error(err, "Failed to process namespaces")

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *ResourceQuotaConfigReconciler) fetchResourceQuotaConfig(ctx context.Context, req ctrl.Request, instance *homelabv1alpha1.ResourceQuotaConfig) error {
	log := log.FromContext(ctx)

	if err := r.Get(ctx, req.NamespacedName, instance); err != nil && !errors.IsNotFound(err) {
		log.Error(err, "Failed to get ResourceQuotaConfig")

		return err
	} else if errors.IsNotFound(err) {
		log.Info("ResourceQuotaConfig not found")

		return nil
	}

	return nil
}

func (r *ResourceQuotaConfigReconciler) fetchNamespaces(ctx context.Context, nsSelector *metav1.LabelSelector, namespaces *corev1.NamespaceList) error {
	log := log.FromContext(ctx)

	listOpts := []client.ListOption{}
	if nsSelector != nil {
		selector, err := metav1.LabelSelectorAsSelector(nsSelector)
		if err != nil {
			log.Error(err, "Failed to create selector from NamespaceSelector")
			return err
		}

		listOpts = append(listOpts, client.MatchingLabelsSelector{Selector: selector})
	}
	if err := r.List(ctx, namespaces, listOpts...); err != nil {
		log.Error(err, "Failed to list Namespaces")

		return err
	}

	return nil
}

func (r *ResourceQuotaConfigReconciler) createResourceQuota(ctx context.Context, ns *corev1.Namespace, instance *homelabv1alpha1.ResourceQuotaConfig) error {
	log := log.FromContext(ctx)

	rq := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *instance.Spec.ResourceQuotaName,
			Namespace: ns.Name,
			Labels:    instance.Spec.ResourceQuotaLabels,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: instance.Spec.ResourceQuotaSpec,
		},
	}

	if err := r.Create(ctx, rq); err != nil {
		log.Error(err, "Failed to create ResourceQuota")

		return err
	}

	return nil
}

func (r *ResourceQuotaConfigReconciler) patchResourceQuota(ctx context.Context, o client.ObjectKey, instance *homelabv1alpha1.ResourceQuotaConfig) error {
	log := log.FromContext(ctx)

	// Fetch the ResourceQuota
	rq := &corev1.ResourceQuota{}
	if err := r.Get(ctx, o, rq); err != nil {
		log.Error(err, "Failed to get ResourceQuota")
		return err
	}

	// Create a patch based on the original ResourceQuota
	patch := client.MergeFrom(rq.DeepCopy())

	// Apply new labels and spec
	rq.Labels = instance.Spec.ResourceQuotaLabels
	rq.Spec.Hard = instance.Spec.ResourceQuotaSpec

	if err := r.Patch(ctx, rq, patch); err != nil {
		log.Error(err, "Failed to patch ResourceQuota")
		return err
	}

	return nil
}

func (r *ResourceQuotaConfigReconciler) containsLabels(selector map[string]string, labels map[string]string) bool {
	// if no selector labels are set, return false as a precaution
	if len(selector) == 0 {
		return false
	}

	// Check if the selector labels are contained in the labels
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}

	return true
}

func (r *ResourceQuotaConfigReconciler) processNamespaces(ctx context.Context, namespaces *corev1.NamespaceList, instance *homelabv1alpha1.ResourceQuotaConfig) error {
	log := log.FromContext(ctx)

	for _, ns := range namespaces.Items {
		// If namespace is not active, skip it
		if ns.Status.Phase != corev1.NamespaceActive {
			continue
		}

		rqList := &corev1.ResourceQuotaList{}
		if err := r.List(ctx, rqList, client.InNamespace(ns.Name)); err != nil {
			log.Error(err, "Failed to list ResourceQuotas")

			return err
		}

		// if no ResourceQuotas exist, create one
		if len(rqList.Items) == 0 {
			if err := r.createResourceQuota(ctx, &ns, instance); err != nil {
				log.Error(err, "Failed to create ResourceQuota")
				continue
			}

			continue
		}

		if len(rqList.Items) == 1 {
			// If only one ResourceQuota exists, check if the labels match
			if r.containsLabels(instance.Spec.ResourceQuotaLabels, rqList.Items[0].Labels) {
				// Create an update request
				object := client.ObjectKey{
					Name:      rqList.Items[0].Name,
					Namespace: rqList.Items[0].Namespace,
				}
				if err := r.patchResourceQuota(ctx, object, instance); err != nil {
					log.Error(err, "Failed to update ResourceQuota")

					return err
				}
			}

			continue
		}

		// We found more than one ResourceQuota. Dafuq? Why do you need more than one ResourceQuota? Just kidding.
		for _, rq := range rqList.Items {
			// Skip ResourceQuotas not managed by the operator
			if !r.containsLabels(instance.Spec.ResourceQuotaLabels, rq.Labels) {
				continue
			}

			// Delete the ResourceQuota since it matches the labels but there is another ResourceQuota not managed by the operator
			if err := r.Delete(ctx, &rq); err != nil {
				log.Error(err, "Failed to delete ResourceQuota")

				return err
			}
		}
	}
	return nil
}

///
///

// findObjectForResourceQuota returns a reconcile request for the ResourceQuotaConfig object
func (r *ResourceQuotaConfigReconciler) findObjectForResourceQuota(ctx context.Context, o client.Object) []reconcile.Request {
	log := log.FromContext(context.Background())
	log.Info("Received request triggered by ResourceQuota creation", "resourcequota", o.GetName())

	// Fetch all ResourceQuotaConfigs, there should only be one
	rqConfigs := &homelabv1alpha1.ResourceQuotaConfigList{}
	if err := r.List(context.Background(), rqConfigs); err != nil {
		log.Error(err, "Failed to list ResourceQuotaConfigs")

		return nil
	}

	// Fetch the first ResourceQuotaConfig
	if len(rqConfigs.Items) == 0 {
		log.Info("No ResourceQuotaConfigs found")

		return nil
	}

	rqConfig := rqConfigs.Items[0]

	// Return a request for the first ResourceQuotaConfig
	return []reconcile.Request{
		{
			NamespacedName: client.ObjectKey{
				Name: rqConfig.Name,
			},
		},
	}
}

// findObjectForNamespace returns a reconcile request for the ResourceQuotaConfig object
func (r *ResourceQuotaConfigReconciler) findObjectForNamespace(ctx context.Context, o client.Object) []reconcile.Request {
	log := log.FromContext(ctx)
	log.Info("Received request triggered by namespace creation", "namespace", o.GetName())

	// Fetch all ResourceQuotaConfigs, there should only be one
	rqConfigs := &homelabv1alpha1.ResourceQuotaConfigList{}
	if err := r.List(ctx, rqConfigs); err != nil {
		log.Error(err, "Failed to list ResourceQuotaConfigs")

		return nil
	}

	// Fetch the first ResourceQuotaConfig
	if len(rqConfigs.Items) == 0 {
		log.Info("No ResourceQuotaConfigs found")

		return nil
	}

	rqConfig := rqConfigs.Items[0]

	// Return a request for the first ResourceQuotaConfig
	return []reconcile.Request{
		{
			NamespacedName: client.ObjectKey{
				Name: rqConfig.Name,
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourceQuotaConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&homelabv1alpha1.ResourceQuotaConfig{}).
		// Here, we need to watch for creation of Namespaces. Other cases should be ignored
		Watches(&corev1.Namespace{}, handler.EnqueueRequestsFromMapFunc(r.findObjectForNamespace), builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				// Namespace creation should also lead to a reconcile:
				// - to check if a ResourceQuota in the same namespace should be created
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return false
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
		})).
		// Here, we need to watch for creation, deletion and updates of ResourceQuotas
		Watches(&corev1.ResourceQuota{}, handler.EnqueueRequestsFromMapFunc(r.findObjectForResourceQuota), builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				// Creation of a ResourceQuota should be reconciled:
				// - to check if a ResourceQuota in the same namespace should be deleted
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				// Deletion of a ResourceQuota should be reconciled:
				// - to check if a ResourceQuota in the same namespace should be created
				return true
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				// Update of a ResourceQuota should be reconciled:
				// - to check if a ResourceQuota must be reconciled
				return true
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
		})).
		Complete(r)
}
