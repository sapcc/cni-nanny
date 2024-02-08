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

package main

import (
	"flag"
	"os"

	"github.com/sapcc/cni-nanny/internal/controller/calico"

	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"

	"github.com/sapcc/cni-nanny/internal/config"
	bgpcontroller "github.com/sapcc/cni-nanny/internal/controller/bgp"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	bgpv1alpha1 "github.com/sapcc/cni-nanny/api/bgp/v1alpha1"
	topologyv1alpha1 "github.com/sapcc/cni-nanny/api/topology/v1alpha1"
	topologycontroller "github.com/sapcc/cni-nanny/internal/controller/topology"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(bgpv1alpha1.AddToScheme(scheme))
	utilruntime.Must(topologyv1alpha1.AddToScheme(scheme))
	utilruntime.Must(v3.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var requeueInterval int
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&config.Cfg.DefaultName, "default-name", "default", "The default resource name.")
	flag.StringVar(&config.Cfg.Namespace, "namespace", "cni-nanny", "The namespace to operate in.")
	flag.StringVar(&config.Cfg.NodeTopologyLabel, "node-topology-label", "topology.kubernetes.io/zone", "The node topology label to handle peer discovery.")
	flag.StringVar(&config.Cfg.JobImageName, "job-image-name", "cni-nanny-discovery", "The name of bgp peer discovery image.")
	flag.StringVar(&config.Cfg.JobImageTag, "job-image-tag", "latest", "The tag of bgp peer discovery image.")
	flag.IntVar(&config.Cfg.BgpRemoteAs, "bgp-remote-as", 12345, "The remote autonomous system of bgp peers.")
	flag.IntVar(&requeueInterval, "requeue-interval", 5, "requeue interval in minutes")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "operator.cninanny.sap.cc",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&bgpcontroller.BgpPeerDiscoveryReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		DefaultName:  config.Cfg.DefaultName,
		Namespace:    config.Cfg.Namespace,
		JobImageName: config.Cfg.JobImageName,
		JobImageTag:  config.Cfg.JobImageTag,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "BgpPeerDiscovery")
		os.Exit(1)
	}
	if err = (&calico.CalicoBgpReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		DefaultName:       config.Cfg.DefaultName,
		Namespace:         config.Cfg.Namespace,
		NodeTopologyLabel: config.Cfg.NodeTopologyLabel,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CalicoBgp")
		os.Exit(1)
	}

	if err = (&topologycontroller.LabelDiscoveryReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		DefaultName:       config.Cfg.DefaultName,
		Namespace:         config.Cfg.Namespace,
		NodeTopologyLabel: config.Cfg.NodeTopologyLabel,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LabelDiscovery")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
