// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var KubeClient *kubernetes.Clientset

func InitializeKubeClient() error {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	kconfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, nil).ClientConfig()
	if err != nil {
		return err
	}
	KubeClient, err = kubernetes.NewForConfig(kconfig)
	if err != nil {
		return err
	}
	return nil
}
