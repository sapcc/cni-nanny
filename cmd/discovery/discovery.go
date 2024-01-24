package main

import (
	"flag"
	bgpv1alpha1 "github.com/sapcc/cni-nanny/api/bgp/v1alpha1"
	topologyv1alpha1 "github.com/sapcc/cni-nanny/api/topology/v1alpha1"
	"github.com/sapcc/cni-nanny/internal/config"
	"github.com/sapcc/cni-nanny/internal/controller/bgp"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme  = runtime.NewScheme()
	discLog = ctrl.Log.WithName("discovery")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(bgpv1alpha1.AddToScheme(scheme))
	utilruntime.Must(topologyv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var probeAddr string
	var requeueInterval int
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.IntVar(&requeueInterval, "requeue-interval", 5, "requeue interval in minutes")
	flag.StringVar(&config.Cfg.DefaultName, "default-name", "default", "The default resource name.")
	flag.StringVar(&config.Cfg.Namespace, "namespace", "kube-system", "The namespace to operate in.")
	flag.StringVar(&config.Cfg.NodeTopologyLabel, "node-topology-label", "", "The node topology label to handle peer discovery.")
	flag.StringVar(&config.Cfg.NodeTopologyValue, "node-topology-value", "", "The node topology value to handle peer discovery.")
	flag.IntVar(&config.Cfg.TraceCount, "traceroute-count", 10, "The count of traceroute packets to send.")
	flag.IntVar(&config.Cfg.BgpNeighborCount, "bgp-neighbor-count", 1, "The count of bgp neighbors.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New())
	discLog.Info("staring discovery worker", "label", config.Cfg.NodeTopologyValue)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                server.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         false,
	})
	if err != nil {
		discLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&bgp.TracerouteDiscoveryReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		discLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		discLog.Error(err, "unable to setup health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		discLog.Error(err, "unable to setup ready check")
		os.Exit(1)
	}

	discLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		discLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
