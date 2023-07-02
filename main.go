package main

import (
	"context"
	"github.com/pixelbender/go-traceroute/traceroute"
	"log"
	"net"
	"sort"
	"time"
)

func main() {
	t := &traceroute.Tracer{
		Config: traceroute.Config{
			Delay:    50 * time.Millisecond,
			Timeout:  time.Second,
			MaxHops:  30,
			Count:    1,
			Networks: []string{"ip4:icmp", "ip4:ip"},
		},
	}
	defer t.Close()
	replies := make(map[int]*traceroute.Reply)
	hops := []int{}
	err := t.Trace(context.Background(), net.ParseIP("1.1.1.1"), func(reply *traceroute.Reply) {
		//log.Printf("%d. %v %v", reply.Hops, reply.IP, reply.RTT)
		replies[reply.Hops] = reply
		hops = append(hops, reply.Hops)
	})
	if err != nil {
		log.Fatal(err)
	}
	sort.Ints(hops)
	for _, hop := range hops {
		reply := replies[hop]
		log.Printf("%d. %v %v", reply.Hops, reply.IP, reply.RTT)
	}
}
