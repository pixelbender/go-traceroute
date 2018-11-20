# go-traceroute

Go implementation of traceroute.

## Features

- [x] Reusable and Concurrent raw IP socket (needs root)
- [x] IPv4 ID and IPv6 FlowLabel tracking
- [ ] IPv6 + ICMP support
- [ ] UDP/TCP support

## Installation

```sh
go get github.com/pixelbender/go-traceroute/...
```

## Example

```go
package main

import (
	"github.com/pixelbender/go-traceroute/traceroute"
	"context"
	"log"
	"net"
	"time"
	"fmt"
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
    ip := net.ParseIP("8.8.8.8")
    err := t.Trace(context.Background(), ip, func(reply *traceroute.Reply) {
        log.Printf("%d. %v %v", reply.Hops, reply.IP, reply.RTT)
    })
    if err != nil {
        log.Fatal(err)
    }
}
```
