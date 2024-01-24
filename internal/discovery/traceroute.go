package discovery

import (
	"context"
	"fmt"
	"github.com/sapcc/go-traceroute/traceroute"
	"net"
	"time"
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
		dst := fmt.Sprintf("1.1.1.%v", i)
		err := t.Trace(context.Background(), net.ParseIP(dst), func(reply *traceroute.Reply) {
			h[reply.IP.String()] = struct{}{}
		})
		if err != nil {
			return nil, err
		}
	}

	var neigh []*net.IP
	for k, _ := range h {
		ip := net.ParseIP(k)
		neigh = append(neigh, &ip)
	}
	return neigh, nil
}
