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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("ResourceQuotaConfig Webhook", func() {

	Context("When creating ResourceQuotaConfig under Defaulting Webhook", func() {
		It("Should fill in the default value if a required field is empty", func() {

			rqc := &ResourceQuotaConfig{
				Spec: ResourceQuotaConfigSpec{},
			}

			rqc.Default()

			Expect(*rqc.Spec.ResourceQuotaName).To(Equal("rq-default"))
			Expect(rqc.Spec.ResourceQuotaLabels).To(Equal(map[string]string{"operator": "rq-operator"}))
			Expect(rqc.Spec.ResourceQuotaSpec).To(Equal(corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			}))
		})

	})

	Context("When creating ResourceQuotaConfig under Validating Webhook", func() {
		It("Should deny if a required field is empty", func() {

			// TODO(user): Add your logic here

		})

		It("Should admit if all required fields are provided", func() {

			// TODO(user): Add your logic here

		})
	})

})
