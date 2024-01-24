package config

import "net"

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
}
