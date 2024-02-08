// Copyright 2024 SAP SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	ServiceAccount    string
	BgpNeighborCount  int
	BgpRemoteAs       int
}
