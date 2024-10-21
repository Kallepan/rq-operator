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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	homelabv1alpha1 "gitlab.srv-lab.server.home/homelab/iac/operators/rq-operator/api/v1alpha1"
)

var _ = Describe("ResourceQuotaConfig Controller without default resource", func() {
	Context("When reconciling a resource", func() {
		ctx := context.Background()

		It("should fail to fetch the ResourceQuotaConfig", func() {
			By("Fetching the ResourceQuotaConfig")

			controllerReconciler := &ResourceQuotaConfigReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			resourcequotaconfig := &homelabv1alpha1.ResourceQuotaConfig{}

			// Test the case where the resource does not exist
			request := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "non-existent-resource",
					Namespace: "default",
				},
			}
			Expect(controllerReconciler.fetchResourceQuotaConfig(ctx, request, resourcequotaconfig)).To(Succeed())
			Expect(resourcequotaconfig.Name).To(BeEmpty())
		})
	})
})

var _ = Describe("ResourceQuotaConfig Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		resourcequotaconfig := &homelabv1alpha1.ResourceQuotaConfig{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ResourceQuotaConfig")
			err := k8sClient.Get(ctx, typeNamespacedName, resourcequotaconfig)
			if err != nil && errors.IsNotFound(err) {
				resource := &homelabv1alpha1.ResourceQuotaConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					// TODO(user): Specify other spec details if needed.
				}
				resource.Default()
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &homelabv1alpha1.ResourceQuotaConfig{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ResourceQuotaConfig")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ResourceQuotaConfigReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})

		It("should successfully fetch the ResourceQuotaConfig", func() {
			By("Fetching the ResourceQuotaConfig")

			controllerReconciler := &ResourceQuotaConfigReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			request := ctrl.Request{
				NamespacedName: typeNamespacedName,
			}
			Expect(controllerReconciler.fetchResourceQuotaConfig(ctx, request, resourcequotaconfig)).To(Succeed())
			Expect(resourcequotaconfig.Name).To(Equal(resourceName))
		})

		It("should successfully fetch the default namespace if no labels are defined", func() {
			By("Fetching the default namespace")

			/// Check with no labels ///
			controllerReconciler := &ResourceQuotaConfigReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			namespaces := &corev1.NamespaceList{}

			if err := controllerReconciler.fetchNamespaces(ctx, nil, namespaces); err != nil {
				Expect(err).NotTo(HaveOccurred())
			}

			Expect(namespaces.Items).NotTo(BeEmpty())
			Expect(namespaces.Items[0].Name).To(Equal("default"))
			Expect(len(namespaces.Items)).To(Equal(4)) // I don't know why it's 4

			/// Check with another namespace ///
			dummyNamespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "dummy",
					Labels: map[string]string{"test": "test"},
				},
			}

			Expect(k8sClient.Create(ctx, dummyNamespace)).To(Succeed())

			if err := controllerReconciler.fetchNamespaces(ctx, nil, namespaces); err != nil {
				Expect(err).NotTo(HaveOccurred())
			}

			Expect(namespaces.Items).NotTo(BeEmpty())
			Expect(len(namespaces.Items)).To(Equal(5))

			/// Check with a label selector ///
			namespaces = &corev1.NamespaceList{}
			if err := controllerReconciler.fetchNamespaces(ctx, &metav1.LabelSelector{
				MatchLabels: map[string]string{"test": "test"},
			}, namespaces); err != nil {
				Expect(err).NotTo(HaveOccurred())
			}

			Expect(namespaces.Items).NotTo(BeEmpty())
			Expect(len(namespaces.Items)).To(Equal(1))
		})

		It("should successfully find the ResourceQuotaConfig for a namespace", func() {
			By("Finding the ResourceQuotaConfig for a namespace")
			controllerReconciler := &ResourceQuotaConfigReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			nsMock := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}

			nsReqs := controllerReconciler.findObjectForResourceQuota(ctx, nsMock)
			Expect(nsReqs).NotTo(BeNil())
			Expect(len(nsReqs)).To(Equal(1))

			By("Finding the ResourceQuotaConfig for a ResourceQuota")
			rqMock := &corev1.ResourceQuota{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-resource",
					Namespace: "default",
				},
			}

			rqReqs := controllerReconciler.findObjectForResourceQuota(ctx, rqMock)
			Expect(rqReqs).NotTo(BeNil())
			Expect(len(rqReqs)).To(Equal(1))
		})

		It("should successfully create a ResourceQuota", func() {
			By("Creating a ResourceQuota")
			controllerReconciler := &ResourceQuotaConfigReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			nsMock := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}
			resourcequotaconfig.Spec.ResourceQuotaName = new(string)
			*resourcequotaconfig.Spec.ResourceQuotaName = "test-resourcequota"
			resourcequotaconfig.Spec.ResourceQuotaSpec = corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			}
			Expect(controllerReconciler.createResourceQuota(ctx, nsMock, resourcequotaconfig)).To(Succeed())
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "test-resourcequota",
				Namespace: "default",
			}, &corev1.ResourceQuota{})).NotTo(HaveOccurred())
		})

		It("should successfully patch a ResourceQuota", func() {
			By("Updating a ResourceQuota")

			controllerReconciler := &ResourceQuotaConfigReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			nsMock := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}

			resourcequotaconfig.Spec.ResourceQuotaName = new(string)
			*resourcequotaconfig.Spec.ResourceQuotaName = "update-resourcequota"
			resourcequotaconfig.Spec.ResourceQuotaSpec = corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			}

			Expect(controllerReconciler.createResourceQuota(ctx, nsMock, resourcequotaconfig)).To(Succeed())
			resourcequotaconfig.Spec.ResourceQuotaSpec = corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			}

			// get the resource quota
			rq := &corev1.ResourceQuota{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "update-resourcequota",
				Namespace: "default",
			}, rq)).NotTo(HaveOccurred())

			// Update the resource quota
			object := client.ObjectKey{
				Name:      "update-resourcequota",
				Namespace: "default",
			}
			Expect(controllerReconciler.patchResourceQuota(ctx, object, resourcequotaconfig)).To(Succeed())

			// get the resource quota
			rq = &corev1.ResourceQuota{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "update-resourcequota",
				Namespace: "default",
			}, rq)).NotTo(HaveOccurred())

			Expect(rq.Spec.Hard).To(Equal(corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("4"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			}))
		})

		It("should check if the labels of the config are in the resource", func() {
			By("Comparing the labels")
			controllerReconciler := &ResourceQuotaConfigReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			tests := []struct {
				selector     map[string]string
				labelsToTest map[string]string
				contains     bool
			}{
				{
					selector: map[string]string{
						"test": "test",
					},
					labelsToTest: map[string]string{
						"test": "test",
					}, contains: true,
				},
				{
					selector: map[string]string{
						"test": "test",
					},
					labelsToTest: map[string]string{},
					contains:     false,
				},
				{
					selector: map[string]string{},
					labelsToTest: map[string]string{
						"test": "test",
					},
					contains: false,
				},
				{
					selector: map[string]string{
						"test": "test",
					},
					labelsToTest: map[string]string{
						"test":  "test",
						"test2": "test2",
					},
					contains: true,
				},
			}

			for i, test := range tests {
				By(fmt.Sprintf("Test %d", i))
				Expect(controllerReconciler.containsLabels(test.selector, test.labelsToTest)).To(Equal(test.contains))
			}
		})
	})
})
