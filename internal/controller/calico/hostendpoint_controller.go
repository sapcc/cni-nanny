// Copyright 2025 SAP SE
// SPDX-License-Identifier: Apache-2.0

package calico

import (
	"context"

	"github.com/go-logr/logr"
	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

type HostEndpointReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	InterfaceName string
}

func (r *HostEndpointReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var hostEndpoint v3.HostEndpoint
	if err := r.Get(ctx, req.NamespacedName, &hostEndpoint); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
	}
	var node corev1.Node
	if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
		return ctrl.Result{}, err
	}
	hostEndpoint.Name = node.Name
	result, err := controllerutil.CreateOrPatch(ctx, r.Client, &hostEndpoint, func() error {
		hostEndpoint.Spec.Node = node.Name
		hostEndpoint.Spec.InterfaceName = r.InterfaceName
		if err := controllerutil.SetOwnerReference(&node, &hostEndpoint, r.Scheme); err != nil {
			return err
		}
		if hostEndpoint.Labels == nil {
			hostEndpoint.Labels = make(map[string]string)
		}
		for key, val := range node.Labels {
			hostEndpoint.Labels[key] = val
		}
		return nil
	})
	if err != nil {
		r.Log.Error(err, "Failed to create or patch HostEndpoint", "name", hostEndpoint.Name)
		return ctrl.Result{}, err
	}
	if result != controllerutil.OperationResultNone {
		r.Log.Info("HostEndpoint reconciled", "operation", result)
	}
	return ctrl.Result{}, nil
}

func (r *HostEndpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v3.HostEndpoint{}).
		Watches(&corev1.Node{}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
