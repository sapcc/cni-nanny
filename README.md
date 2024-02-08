# cni-nanny

[![CI](https://github.com/sapcc/cni-nanny/actions/workflows/ci.yaml/badge.svg)](https://github.com/sapcc/cni-nanny/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sapcc/cni-nanny)](https://goreportcard.com/report/github.com/sapcc/cni-nanny)

Helper for CNI operations. Set of controllers help to discover and configure Calico BGP peers.  

[![architecture](docs/images/architecture.png)]

Why?
----

Each rack TORs have different IP addresses for peering with Calico nodes. Peer discovery is based on traceroute. 
