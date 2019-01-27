# go-traceroute

Go implementation of traceroute.

## Features

- [x] Reusable and Concurrent raw IP socket (needs root)
- [x] IPv4 ID and IPv6 FlowLabel tracking
- [ ] ICMPv6 support
- [ ] UDP/TCP support

## Installation

```sh
go get github.com/pixelbender/go-traceroute/...
```

## Simple tracing

```go
package main

import (
	"github.com/pixelbender/go-traceroute/traceroute"
	"log"
)

func main() {
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
```

## Custom configuration

```go
package main

import (
	"github.com/pixelbender/go-traceroute/traceroute"
	"context"
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
```
