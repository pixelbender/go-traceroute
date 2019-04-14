package traceroute_test

import (
	"context"
	"github.com/pixelbender/go-traceroute/traceroute"
	"net"
	"testing"
	"time"
)

func TestTraceReply(t *testing.T) {
	err := traceroute.DefaultTracer.Trace(
		context.Background(),
		net.ParseIP("1.1.1.1"),
		func(r *traceroute.Reply) {
			t.Logf("%d. %v %v", r.Hops, r.IP, r.RTT)
		})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPing(t *testing.T) {
	sess, err := traceroute.NewSession(net.ParseIP("1.1.1.1"))
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	err = sess.Ping(traceroute.DefaultConfig.MaxHops)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case r := <-sess.Receive():
		t.Logf("%v %v", r.IP, r.RTT)
	case <-time.After(traceroute.DefaultConfig.Timeout):
		t.Fatal("time out")
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
