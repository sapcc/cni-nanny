package config

import "net"

const (
	KubeApp            = "cni-nanny"
	KubeLabelComponent = "app.kubernetes.io/component"
	KubeLabelManaged   = "app.kubernetes.io/managed-by"
)

var Cfg = Config{}

type Config struct {
	DefaultName       string
	Namespace         string
	NodeTopologyLabel string
	NodeTopologyValue string
	TraceCount        int
	StartingIP        net.IP
	JobImageName      string
	JobImageTag       string
	BgpNeighborCount  int
	BgpRemoteAs       int
}
