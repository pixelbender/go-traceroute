package traceroute_test

import (
	"context"
	"github.com/pixelbender/go-traceroute/traceroute"
	"net"
	"testing"
	"time"
)

func TestTraceReply(t *testing.T) {
	ip := net.ParseIP("1.1.1.1")
	err := traceroute.DefaultTracer.Trace(context.Background(), ip, func(reply *traceroute.Reply) {
		t.Logf("%d. %v %v", reply.Hops, reply.IP, reply.RTT)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTrace(t *testing.T) {
	ip := net.ParseIP("1.1.1.1")
	hops, err := traceroute.Trace(ip)
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range hops {
		for _, n := range h.Nodes {
			t.Logf("%d. %v %v", h.Distance, n.IP, n.RTT)
		}
	}
}

func TestConcurrent(t *testing.T) {
	hosts := []string{
		"1.1.1.1",
		"8.8.8.8",
	}
	ch := make(chan []*traceroute.Hop, len(hosts))
	for _, h := range hosts {
		go func(host string) {
			ip := net.ParseIP(host)
			hops, err := traceroute.Trace(ip)
			if err != nil {
				t.Log(err)
			}
			ch <- hops
		}(h)
	}

	done := 0
	for {
		select {
		case hops := <-ch:
			for _, h := range hops {
				for _, n := range h.Nodes {
					t.Logf("%d. %v %v", h.Distance, n.IP, n.RTT)
				}
			}
			if done++; done == len(hosts) {
				return
			}
		case <-time.After(10 * time.Second):
			t.Fatal("timeout")
		}
	}
}
