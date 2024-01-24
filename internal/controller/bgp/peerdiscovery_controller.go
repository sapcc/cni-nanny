package bgp

import (
	"context"
	bgpv1alpha1 "github.com/sapcc/cni-nanny/api/bgp/v1alpha1"
	topologyv1alpha1 "github.com/sapcc/cni-nanny/api/topology/v1alpha1"
	"github.com/sapcc/cni-nanny/internal/config"
	"github.com/sapcc/cni-nanny/internal/discovery"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type TracerouteDiscoveryReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	DefaultName       string
	Namespace         string
	NodeTopologyLabel string
	NodeTopologyValue string
	TraceCount        int
	BgpNeighborCount  int
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BgpPeerDiscovery object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *TracerouteDiscoveryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	peers, err := discovery.GetNeighbors(config.Cfg.TraceCount)
	if err != nil {
		log.FromContext(ctx).Error(err, "unable to discover peers")
		os.Exit(1)
	}
	log.FromContext(ctx).Info("peers found", "peers", peers)

	var bgpPeerDiscovery = new(bgpv1alpha1.BgpPeerDiscovery)
	var nsName types.NamespacedName
	var peerList []string

	if len(peers) > 0 {
		for _, v := range peers {
			peerList = append(peerList, v.String())
		}
		nsName.Name = config.Cfg.NodeTopologyValue
		nsName.Namespace = config.Cfg.Namespace
		err := r.Get(ctx, nsName, bgpPeerDiscovery)
		if err != nil {
			if errors.IsNotFound(err) {
				labelDisc, err := generateBgpPeerDiscovery(nsName, bgpPeerDiscovery)
				log.FromContext(ctx).Info("here", "name", req.Name)
				if err != nil {
					log.FromContext(ctx).Error(err, "error generating bgpPeerDiscovery")
					return ctrl.Result{}, err
				}
				err = r.Create(ctx, &labelDisc)
				if err != nil {
					log.FromContext(ctx).Error(err, "error creating bgpPeerDiscovery")
					return ctrl.Result{}, err
				}
			}
			log.FromContext(ctx).Error(err, "error getting bgpPeerDiscovery")
			return ctrl.Result{}, err
		}
		if bgpPeerDiscovery.Status.DiscoveredPeers == nil {
			patch := client.MergeFrom(bgpPeerDiscovery.DeepCopy())
			err := r.updateStatus(ctx, peerList, patch, bgpPeerDiscovery)
			if err != nil {
				log.FromContext(ctx).Error(err, "error updating bgpPeerDiscovery status")
			}
		}
	}

	os.Exit(0)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TracerouteDiscoveryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("peer discovery controller").
		For(&topologyv1alpha1.LabelDiscovery{}).
		Complete(r)
}

func generateBgpPeerDiscovery(nsName types.NamespacedName, bgpPeerDiscovery *bgpv1alpha1.BgpPeerDiscovery) (bgpv1alpha1.BgpPeerDiscovery, error) {
	bgpPeerDiscovery.Name = nsName.Name
	bgpPeerDiscovery.Namespace = nsName.Namespace
	//bgpPeerDiscovery.Status = make
	//labeldisc.ObjectMeta.Labels = map[string]string{}
	//labeldisc.ObjectMeta.Labels[] =
	return *bgpPeerDiscovery, nil
}

func (r *TracerouteDiscoveryReconciler) updateStatus(ctx context.Context, peers []string, patch client.Patch, bgpPeerDiscovery *bgpv1alpha1.BgpPeerDiscovery) error {
	bgpPeerDiscovery.Status.DiscoveredPeers = peers
	err := r.Status().Patch(ctx, bgpPeerDiscovery, patch)
	if err != nil {
		return err
	}
	return nil
}
