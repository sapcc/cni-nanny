// Copyright 2025 SAP SE
// SPDX-License-Identifier: Apache-2.0

package calico_test

import (
	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HostEndpointReconciler", func() {

	var node *corev1.Node

	BeforeEach(func(ctx SpecContext) {
		node = &corev1.Node{}
		node.Name = "test-node"
		node.Labels = map[string]string{
			"topology.kubernetes.io/zone": "test-zone",
		}
		Expect(k8sClient.Create(ctx, node)).To(Succeed())
	})

	AfterEach(func(ctx SpecContext) {
		Expect(k8sClient.Delete(ctx, node)).To(Succeed())
	})

	It("creates a HostEndpoint", func(ctx SpecContext) {
		var hostEndpoint v3.HostEndpoint
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Name: node.Name}, &hostEndpoint)
		}).Should(Succeed())
		Expect(hostEndpoint.Name).To(Equal(node.Name))
		Expect(hostEndpoint.Spec.Node).To(Equal(node.Name))
		Expect(hostEndpoint.Spec.InterfaceName).To(Equal("the-interface"))
	})

	It("copies node labels to HostEndpoint", func(ctx SpecContext) {
		var hostEndpoint v3.HostEndpoint
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Name: node.Name}, &hostEndpoint)
		}).Should(Succeed())
		Expect(hostEndpoint.Labels).To(HaveKeyWithValue("topology.kubernetes.io/zone", "test-zone"))
	})

	It("sets an owner reference on the HostEndpoint", func(ctx SpecContext) {
		var hostEndpoint v3.HostEndpoint
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Name: node.Name}, &hostEndpoint)
		}).Should(Succeed())
		Expect(hostEndpoint.OwnerReferences).To(HaveLen(1))
		ownerRef := hostEndpoint.OwnerReferences[0]
		Expect(ownerRef.Name).To(Equal(node.Name))
		Expect(ownerRef.Kind).To(Equal("Node"))
		Expect(ownerRef.APIVersion).To(Equal("v1"))
	})

})
