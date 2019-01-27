package traceroute_test

import (
	"context"
	"github.com/pixelbender/go-traceroute/traceroute"
	"log"
	"net"
	"testing"
	"time"
)

func TestTraceReply(t *testing.T) {
	ip := net.ParseIP("8.8.8.8")
	err := traceroute.DefaultTracer.Trace(context.Background(), ip, func(reply *traceroute.Reply) {
		t.Logf("%d. %v %v", reply.Hops, reply.IP, reply.RTT)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTrace(t *testing.T) {
	ip := net.ParseIP("8.8.8.8")
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

func ExampleSimple() {
	hops, err := traceroute.Trace(net.ParseIP("8.8.8.8"))
	if err != nil {
		log.Fatal(err)
	}
	for _, h := range hops {
		for _, n := range h.Nodes {
			log.Printf("%d. %v %v", h.Distance, n.IP, n.RTT)
		}
	}
}

func ExampleCustom() {
	t := &traceroute.Tracer{
		Config: traceroute.Config{
			Delay:   50 * time.Millisecond,
			Timeout: time.Second,
			MaxHops: 30,
			Count:   3,
			Network: "ip4:ip",
		},
	}
	defer t.Close()
	err := t.Trace(context.Background(), net.ParseIP("8.8.8.8"), func(reply *traceroute.Reply) {
		log.Printf("%d. %v %v", reply.Hops, reply.IP, reply.RTT)
	})
	if err != nil {
		log.Fatal(err)
	}
}
