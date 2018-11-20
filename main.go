package main

import (
	"context"
	"github.com/pixelbender/go-traceroute/traceroute"
	"log"
	"net"
	"time"
)

func main() {
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