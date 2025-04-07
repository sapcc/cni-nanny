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

package bgp

import (
	"context"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	topologyv1alpha1 "github.com/sapcc/cni-nanny/api/topology/v1alpha1"
	"github.com/sapcc/cni-nanny/internal/config"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// BgpPeerDiscoveryReconciler reconciles a BgpPeerDiscovery object
type BgpPeerDiscoveryReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	DefaultName     string
	Namespace       string
	JobImageName    string
	JobImageTag     string
	ServiceAccount  string
	RequeueInterval time.Duration
}

//+kubebuilder:rbac:groups=bgp.cninanny.sap.cc,resources=bgppeerdiscoveries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bgp.cninanny.sap.cc,resources=bgppeerdiscoveries/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bgp.cninanny.sap.cc,resources=bgppeerdiscoveries/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BgpPeerDiscovery object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *BgpPeerDiscoveryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	labelDiscovery := &topologyv1alpha1.LabelDiscovery{}
	err := r.Get(ctx, req.NamespacedName, labelDiscovery)
	if err != nil {
		log.FromContext(ctx).Error(err, "could not get client from manager")
		return reconcile.Result{}, err
	}
	log.FromContext(ctx).Info("found labeldiscovery", labelDiscovery.Name, labelDiscovery.Namespace)
	for k, v := range labelDiscovery.Status.DiscoveredTopologyValues {
		if !v.Finalized {
			log.FromContext(ctx).Info("found non finalized", "topology value", k)
			// check if discovery is already running
			found, err := r.checkJobsForTopologyValue(ctx, k)
			if err != nil {
				log.FromContext(ctx).Error(err, "error checking discovery jobs")
				return ctrl.Result{}, err
			}
			log.FromContext(ctx).Info(k, "existing job found:", found)
			if !found {
				log.FromContext(ctx).Info("starting discovery job", "topology value", k)
				conf := config.Config{
					Namespace:         req.Namespace,
					JobImageName:      r.JobImageName,
					JobImageTag:       r.JobImageTag,
					NodeTopologyLabel: labelDiscovery.Spec.TopologyLabel,
					NodeTopologyValue: k,
					ServiceAccount:    r.ServiceAccount,
				}
				err = r.createDiscoveryJob(ctx, conf)
				if err != nil {
					log.FromContext(ctx).Error(err, "error creating job")
					return ctrl.Result{}, err
				}
			}
		}
	}
	return ctrl.Result{
		RequeueAfter: r.RequeueInterval,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BgpPeerDiscoveryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("bgp peer discovery controller").
		For(&topologyv1alpha1.LabelDiscovery{}).
		Complete(r)
}

func (r BgpPeerDiscoveryReconciler) checkJobsForTopologyValue(ctx context.Context, value string) (bool, error) {
	labelSelector, err := labels.ValidatedSelectorFromSet(map[string]string{topologyv1alpha1.TopologyValue: value})
	if err != nil {
		log.FromContext(ctx).Error(err, "error building label selector")
		return false, err
	}
	listOption := client.ListOptions{LabelSelector: labelSelector}
	jobList := batchv1.JobList{}
	err = r.List(ctx, &jobList, &listOption)
	if err != nil {
		log.FromContext(ctx).Error(err, "error getting jobs")
		return false, err
	}
	if len(jobList.Items) == 0 {
		log.FromContext(ctx).Info("no jobs found")
		return false, nil
	}
	return true, nil
}

func (r BgpPeerDiscoveryReconciler) createDiscoveryJob(ctx context.Context, conf config.Config) error {
	job := batchv1.Job{Spec: batchv1.JobSpec{}}
	lab := map[string]string{}
	lab[config.KubeLabelComponent] = "DiscoveryJob"
	lab[config.KubeLabelManaged] = config.KubeApp
	lab[topologyv1alpha1.TopologyValue] = conf.NodeTopologyValue
	job.Name = "bgp-peer-discovery" + "-" + conf.NodeTopologyValue
	job.Namespace = conf.Namespace
	job.Labels = lab

	timeToLive := int32(60)
	sel := make(map[string]string)
	sel[conf.NodeTopologyLabel] = conf.NodeTopologyValue

	job.Spec.TTLSecondsAfterFinished = &timeToLive

	job.Spec.Template = corev1.PodTemplateSpec{}
	job.Spec.Template.Spec = corev1.PodSpec{}
	job.Spec.Template.Labels = lab
	job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever
	job.Spec.Template.Spec.NodeSelector = sel
	job.Spec.Template.Spec.HostNetwork = true
	job.Spec.Template.Spec.ServiceAccountName = conf.ServiceAccount
	job.Spec.Template.Spec.Tolerations = []corev1.Toleration{
		{
			Operator: corev1.TolerationOpExists,
		},
	}
	container := corev1.Container{
		Image: conf.JobImageName + ":" + conf.JobImageTag,
		Name:  "discover",
		Args: []string{
			"--node-topology-label", conf.NodeTopologyLabel,
			"--node-topology-value", conf.NodeTopologyValue,
		},
	}
	job.Spec.Template.Spec.Containers = []corev1.Container{container}
	err := r.Create(ctx, &job)
	if err != nil {
		return err
	}
	return nil
}
