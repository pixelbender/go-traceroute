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
			Count:    3,
			Networks: []string{"ip4:icmp", "ip4:ip"},
		},
	}
	defer t.Close()
	replies := make(map[int][]*traceroute.Reply)
	err := t.Trace(context.Background(), net.ParseIP("1.1.1.1"), func(reply *traceroute.Reply) {
		//log.Printf("%d. %v %v", reply.Hops, reply.IP, reply.RTT)
		replies[reply.Hops] = append(replies[reply.Hops], reply)
	})
	if err != nil {
		log.Fatal(err)
	}

	printHops(replies)
}

func printHops(replies map[int][]*traceroute.Reply) {
	hops := []int{}
	for hop := range replies {
		hops = append(hops, hop)
	}
	sort.Ints(hops)
	for _, hop := range hops {
		replyList := replies[hop]
		for _, reply := range replyList {
			log.Printf("%d. %v %v", reply.Hops, reply.IP, reply.RTT)
		}
	}
}
