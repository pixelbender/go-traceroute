package traceroute

import (
	"context"
	"net"
	"testing"
)

func TestTraceReply(t *testing.T) {
	ip := net.ParseIP("8.8.8.8")
	err := DefaultTracer.Trace(context.Background(), ip, func(reply *Reply) {
		t.Logf("%d. %v %v", reply.Hops, reply.IP, reply.RTT)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTrace(t *testing.T) {
	ip := net.ParseIP("8.8.8.8")
	hops, err := Trace(ip)
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range hops {
		for _, n := range h.Nodes {
			t.Logf("%d. %v %v", h.Distance, n.IP, n.RTT)
		}
	}
}
