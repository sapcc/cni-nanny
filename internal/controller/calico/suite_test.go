// Copyright 2025 SAP SE
// SPDX-License-Identifier: Apache-2.0

package calico_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/sapcc/cni-nanny/internal/controller/calico"
)

var (
	testEnv        *envtest.Environment
	k8sClient      client.Client
	k8sManager     ctrl.Manager
	stopController context.CancelFunc
)

func TestMirror(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(4 * time.Second)
	RunSpecs(t, "Calico Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}
	testEnv.CRDInstallOptions.Paths = []string{"crd"}
	testEnv.CRDInstallOptions.ErrorIfPathMissing = true

	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = corev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = v3.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
		Logger: GinkgoLogr,
	})
	Expect(err).ToNot(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	hostEndpointReconciler := calico.HostEndpointReconciler{
		Client:        k8sClient,
		Log:           GinkgoLogr,
		Scheme:        scheme.Scheme,
		InterfaceName: "the-interface",
	}
	Expect(hostEndpointReconciler.SetupWithManager(k8sManager)).To(Succeed())

	go func() {
		defer GinkgoRecover()
		stopCtx, cancel := context.WithCancel(ctrl.SetupSignalHandler())
		stopController = cancel
		err = k8sManager.Start(stopCtx)
		Expect(err).ToNot(HaveOccurred())
	}()

	// The default namespace needs some time to be created...
	time.Sleep(20 * time.Millisecond)
})

var _ = AfterSuite(func() {
	stopController()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
