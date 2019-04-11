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
	err := traceroute.DefaultTracer.Trace(context.Background(), ip, func(r *traceroute.Reply) {
		t.Logf("%d. %v %v", r.Hops, r.IP, r.RTT)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPing(t *testing.T) {
	ip := net.ParseIP("1.1.1.1")
	err := traceroute.DefaultTracer.Ping(context.Background(), ip, func(r *traceroute.Reply) {
		t.Logf("%v %v", r.IP, r.RTT)
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
	ip := net.ParseIP("1.1.1.1")
	count := 3

	ch := make(chan []*traceroute.Hop, count)
	for i := 0; i < count; i++ {
		go func() {
			hops, err := traceroute.Trace(ip)
			if err != nil {
				t.Log(err)
			}
			ch <- hops
		}()
	}
	for count > 0 {
		select {
		case hops := <-ch:
			count--
			for _, h := range hops {
				for _, n := range h.Nodes {
					t.Logf("%d. %v %v", h.Distance, n.IP, n.RTT)
				}
			}
		case <-time.After(10 * time.Second):
			t.Fatal("timeout")
		}
	}
}

func TestRepeat(t *testing.T) {
	tracer := &traceroute.Tracer{
		Config: traceroute.DefaultConfig,
	}
	tracer.Count = 1

	ip := net.ParseIP("1.1.1.1")
	count := 3

	handle := func(r *traceroute.Reply) {
		t.Log(r.Hops, r.IP, r.RTT)
	}
	ctx := context.Background()

	for i := 0; i < count; i++ {
		err := tracer.Trace(ctx, ip, handle)
		if err != nil {
			t.Fatal(err)
		}
	}
}
