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

package calico

import (
	"context"
	"fmt"

	"errors"

	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	"github.com/projectcalico/api/pkg/lib/numorstring"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	bgpv1alpha1 "github.com/sapcc/cni-nanny/api/bgp/v1alpha1"
	topologyv1alpha1 "github.com/sapcc/cni-nanny/api/topology/v1alpha1"
	"github.com/sapcc/cni-nanny/internal/config"
)

// CalicoBgpReconciler reconciles a BgpPeerDiscovery object
type CalicoBgpReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	DefaultName       string
	Namespace         string
	NodeTopologyLabel string
	BgpRemoteAs       int
}

//+kubebuilder:rbac:groups=bgp.cninanny.sap.cc,resources=bgppeerdiscoveries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bgp.cninanny.sap.cc,resources=bgppeerdiscoveries/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bgp.cninanny.sap.cc,resources=bgppeerdiscoveries/finalizers,verbs=update
//+kubebuilder:rbac:groups=topology.cninanny.sap.cc,resources=labeldiscoveries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=topology.cninanny.sap.cc,resources=labeldiscoveries/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=topology.cninanny.sap.cc,resources=labeldiscoveries/finalizers,verbs=update
//+kubebuilder:rbac:groups=projectcalico.org,resources=bgppeers,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BgpPeerDiscovery object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *CalicoBgpReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var bgpPeerDiscovery = new(bgpv1alpha1.BgpPeerDiscovery)
	var nsName types.NamespacedName

	err := r.Get(ctx, req.NamespacedName, bgpPeerDiscovery)
	if err != nil {
		log.FromContext(ctx).Error(err, "error getting bgpPeerDiscovery")
		return ctrl.Result{}, err
	}
	if len(bgpPeerDiscovery.Status.DiscoveredPeers) > 0 {
		for _, v := range bgpPeerDiscovery.Status.DiscoveredPeers {
			var calicoBgpPeer v3.BGPPeer
			nsName.Name = "bgp-peer-" + req.Name + "-" + v
			nsName.Namespace = config.Cfg.Namespace
			asNumber, err := intToUint32(config.Cfg.BgpRemoteAs)
			if err != nil {
				log.FromContext(ctx).Error(err, "error converting BgpRemoteAs to uint32")
				return ctrl.Result{}, err
			}
			spec := v3.BGPPeerSpec{
				PeerIP:       v,
				ASNumber:     numorstring.ASNumber(asNumber),
				NodeSelector: config.Cfg.NodeTopologyLabel + " == " + fmt.Sprintf("%q", req.Name),
			}
			if len(config.Cfg.BgpFilters) > 0 {
				spec.Filters = config.Cfg.BgpFilters
			}
			err = r.Get(ctx, nsName, &calicoBgpPeer)
			if err != nil {
				if k8serrors.IsNotFound(err) {
					calicoPeer := generateCalicoBgpPeer(nsName, spec, &calicoBgpPeer)
					log.FromContext(ctx).Info("creating calico peer", calicoPeer.Name, calicoPeer.Spec.PeerIP)
					err = r.Create(ctx, calicoPeer)
					if err != nil && !k8serrors.IsAlreadyExists(err) {
						log.FromContext(ctx).Error(err, "error creating calicoBgpPeer")
						return ctrl.Result{}, err
					}
				} else {
					log.FromContext(ctx).Error(err, "error getting calicoBgpPeer")
					return ctrl.Result{}, err
				}
			}
		}

		labelDiscovery := &topologyv1alpha1.LabelDiscovery{}
		nsName.Name = config.Cfg.DefaultName
		nsName.Namespace = config.Cfg.Namespace
		err := r.Get(ctx, nsName, labelDiscovery)
		if err != nil {
			log.FromContext(ctx).Error(err, "could not get client from manager")
			return reconcile.Result{}, err
		}
		patch := client.MergeFrom(labelDiscovery.DeepCopy())
		ls := labelDiscovery.Status.DiscoveredTopologyValues
		for k, v := range labelDiscovery.Status.DiscoveredTopologyValues {
			ls[k] = v
			if k == req.Name && !v.Finalized {
				v.Finalized = true
				ls[k] = v
			}
		}
		labelDiscovery.Status.DiscoveredTopologyValues = ls
		err = r.Status().Patch(ctx, labelDiscovery, patch)
		if err != nil {
			log.FromContext(ctx).Error(err, "could not patch labelDiscovery")
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CalicoBgpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("calico bgp controller").
		For(&bgpv1alpha1.BgpPeerDiscovery{}).
		Complete(r)
}

func intToUint32(value int) (uint32, error) {
	if value < 0 || value > int(^uint32(0)) {
		return 0, errors.New("integer overflow: value out of range for uint32")
	}
	return uint32(value), nil
}

func generateCalicoBgpPeer(nsName types.NamespacedName, spec v3.BGPPeerSpec, calicoBgpPeer *v3.BGPPeer) *v3.BGPPeer {
	calicoBgpPeer.Name = nsName.Name
	calicoBgpPeer.Namespace = nsName.Namespace
	calicoBgpPeer.Spec = spec
	calicoBgpPeer.ObjectMeta.Labels = map[string]string{}
	calicoBgpPeer.ObjectMeta.Labels[config.KubeLabelComponent] = "BgpPeer"
	calicoBgpPeer.ObjectMeta.Labels[config.KubeLabelManaged] = config.KubeApp
	return calicoBgpPeer
}
