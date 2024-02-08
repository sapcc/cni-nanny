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

package topology

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	topologyv1alpha1 "github.com/sapcc/cni-nanny/api/topology/v1alpha1"
	"github.com/sapcc/cni-nanny/internal/config"
)

// LabelDiscoveryReconciler reconciles a LabelDiscovery object
type LabelDiscoveryReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	DefaultName       string
	Namespace         string
	NodeTopologyLabel string
}

//+kubebuilder:rbac:groups=topology.cninanny.sap.cc,resources=labeldiscoveries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=topology.cninanny.sap.cc,resources=labeldiscoveries/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=topology.cninanny.sap.cc,resources=labeldiscoveries/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LabelDiscovery object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *LabelDiscoveryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	var labelDiscovery = new(topologyv1alpha1.LabelDiscovery)
	var nsName types.NamespacedName

	nsName.Name = r.DefaultName
	nsName.Namespace = r.Namespace
	err := r.Get(ctx, nsName, labelDiscovery)
	if err != nil {
		if errors.IsNotFound(err) {
			labelDisc := generateLabelDiscovery(nsName, r.NodeTopologyLabel, labelDiscovery)
			err = r.Create(ctx, &labelDisc)
			if err != nil {
				log.FromContext(ctx).Error(err, "error creating labelDiscovery")
				return ctrl.Result{}, err
			} else {
				log.FromContext(ctx).Error(err, "error getting labelDiscovery")
				return ctrl.Result{}, err
			}
		}
	}
	node := &corev1.Node{}
	err = r.Get(ctx, req.NamespacedName, node)
	if err != nil {
		log.FromContext(ctx).Error(err, "could not get client from manager")
		return reconcile.Result{}, err
	}
	var val topologyv1alpha1.DiscoveredTopologyValue

	labelName := node.Labels[r.NodeTopologyLabel]
	val.Finalized = false

	if !containsKey(labelDiscovery.Status.DiscoveredTopologyValues, labelName) {
		patch := client.MergeFrom(labelDiscovery.DeepCopy())
		log.FromContext(ctx).Info("appending label value", labelDiscovery.Name, labelName)
		err := r.appendStatus(ctx, labelName, &val, patch, labelDiscovery)
		if err != nil {
			log.FromContext(ctx).Error(err, "error updating labelDiscovery status")
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LabelDiscoveryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("topology controller").
		For(&corev1.Node{}).
		Complete(r)
}

func generateLabelDiscovery(nsName types.NamespacedName, label string, labelDiscovery *topologyv1alpha1.LabelDiscovery) topologyv1alpha1.LabelDiscovery {
	spec := topologyv1alpha1.LabelDiscoverySpec{
		TopologyLabel: label,
	}
	labelDiscovery.Spec = spec
	labelDiscovery.Name = nsName.Name
	labelDiscovery.Namespace = nsName.Namespace
	labelDiscovery.Status.DiscoveredTopologyValues = make(map[string]topologyv1alpha1.DiscoveredTopologyValue)
	labelDiscovery.ObjectMeta.Labels = map[string]string{}
	labelDiscovery.ObjectMeta.Labels[config.KubeLabelComponent] = "LabelDiscovery"
	labelDiscovery.ObjectMeta.Labels[config.KubeLabelManaged] = config.KubeApp
	return *labelDiscovery
}

func (r *LabelDiscoveryReconciler) appendStatus(ctx context.Context, name string, label *topologyv1alpha1.DiscoveredTopologyValue, patch client.Patch, labelDiscovery *topologyv1alpha1.LabelDiscovery) error {
	if labelDiscovery.Status.DiscoveredTopologyValues == nil {
		labelDiscovery.Status.DiscoveredTopologyValues = make(map[string]topologyv1alpha1.DiscoveredTopologyValue)
	}
	labelDiscovery.Status.DiscoveredTopologyValues[name] = *label
	err := r.Status().Patch(ctx, labelDiscovery, patch)
	if err != nil {
		return err
	}
	return nil
}

func containsKey[M ~map[K]V, K comparable, V any](m M, k K) bool {
	_, ok := m[k]
	return ok
}
