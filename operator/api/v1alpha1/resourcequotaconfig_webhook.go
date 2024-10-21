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

package v1alpha1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var resourcequotaconfiglog = log.Log.WithName("resourcequotaconfig-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *ResourceQuotaConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithValidator(&resourceQuotaConfigValidator{client: mgr.GetClient()}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-homelab-server-home-v1alpha1-resourcequotaconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=homelab.server.home,resources=resourcequotaconfigs,verbs=create;update,versions=v1alpha1,name=mresourcequotaconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ResourceQuotaConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ResourceQuotaConfig) Default() {
	resourcequotaconfiglog.Info("default", "name", r.Name)
	if r.Spec.ResourceQuotaName == nil {
		r.Spec.ResourceQuotaName = new(string)
		*r.Spec.ResourceQuotaName = "rq-default"
	}

	if r.Spec.ResourceQuotaLabels == nil {
		r.Spec.ResourceQuotaLabels = map[string]string{"operator": "rq-operator"}
	}

	if r.Spec.ResourceQuotaSpec == nil {
		r.Spec.ResourceQuotaSpec = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-homelab-server-home-v1alpha1-resourcequotaconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=homelab.server.home,resources=resourcequotaconfigs,verbs=create;update,versions=v1alpha1,name=vresourcequotaconfig.kb.io,admissionReviewVersions=v1

type resourceQuotaConfigValidator struct {
	client client.Client
}

// ValidateCreate implements admission.CustomValidator.
// Check if ResourceQuotaConfig exists. If so, do not allow creation.
func (v *resourceQuotaConfigValidator) ValidateCreate(ctx context.Context, o runtime.Object) (admission.Warnings, error) {
	log := log.FromContext(ctx)
	log.Info("ValidateCreate called")

	// use object to fetch the list of ResourceQuotaConfig
	rqcList := &ResourceQuotaConfigList{}
	if err := v.client.List(ctx, rqcList); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	if len(rqcList.Items) > 0 {
		return admission.Warnings{"ResourceQuotaConfig already exists"}, errors.NewBadRequest("ResourceQuotaConfig already exists")
	}

	return nil, nil
}

// ValidateDelete implements admission.CustomValidator.
func (v *resourceQuotaConfigValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	return nil, nil
}

// ValidateUpdate implements admission.CustomValidator.
func (v *resourceQuotaConfigValidator) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (warnings admission.Warnings, err error) {
	return nil, nil
}
