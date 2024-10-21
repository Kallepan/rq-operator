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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ResourceQuotaConfigSpec defines the desired state of ResourceQuotaConfig
type ResourceQuotaConfigSpec struct {
	// Name to apply to the resource quota
	// By default it will be "rq-default"
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ResourceQuotaName *string `json:"resourceQuotaName,omitempty"`

	// Labels to apply to the resource quota
	// By default it will be "operator: rq-operator"
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ResourceQuotaLabels map[string]string `json:"resourceQuotaLabels,omitempty"`

	// Labels for namespace selection by the operator to apply resource quotas.
	// By default (nil) it will match all namespaces. (This may not be a good idea)
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`

	// Resourcequota spec to apply by default
	// By default it will apply cpu: 2, memory: 2Gi
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ResourceQuotaSpec corev1.ResourceList `json:"resourceQuotaSpec,omitempty"`
}

// ResourceQuotaConfigStatus defines the observed state of ResourceQuotaConfig
type ResourceQuotaConfigStatus struct {
	// Represents the latest available observations of a resource's current state.
	// ResourceQuotaConfig.status.conditions.type are: "Ready", "Progressing", "Degraded"
	// ResourceQuotaConfig.status.conditions.status are: "True", "False", "Unknown"
	// ResourceQuotaConfig.status.conditions.reason is a brief machine-readable explanation for the condition's last transition.
	// ResourceQuotaConfig.status.conditions.Message is a human-readable message indicating details about the transition.

	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ResourceQuotaConfig is the Schema for the resourcequotaconfigs API
type ResourceQuotaConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceQuotaConfigSpec   `json:"spec,omitempty"`
	Status ResourceQuotaConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ResourceQuotaConfigList contains a list of ResourceQuotaConfig
type ResourceQuotaConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceQuotaConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceQuotaConfig{}, &ResourceQuotaConfigList{})
}
