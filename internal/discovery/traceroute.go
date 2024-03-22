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

package discovery

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/sapcc/go-traceroute/traceroute"
)

// GetNeighbors discovers next-hops by sending traceroute packets with ttl=1
func GetNeighbors(count int) ([]*net.IP, error) {
	t := &traceroute.Tracer{
		Config: traceroute.Config{
			Delay:    50 * time.Millisecond,
			Timeout:  time.Second,
			MaxHops:  1,
			Count:    1,
			Networks: []string{"ip4:icmp", "ip4:ip"},
		},
	}
	defer t.Close()

	h := make(map[string]struct{})
	for i := 0; i < count; i++ {
		dst := fmt.Sprintf("8.8.8.%v", i)
		err := t.Trace(context.Background(), net.ParseIP(dst), func(reply *traceroute.Reply) {
			h[reply.IP.String()] = struct{}{}
		})
		if err != nil {
			return nil, err
		}
	}

	var neigh []*net.IP
	for k := range h {
		ip := net.ParseIP(k)
		neigh = append(neigh, &ip)
	}
	return neigh, nil
}
