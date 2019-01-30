# go-traceroute

Go implementation of traceroute.

[![Build Status](https://api.travis-ci.org/pixelbender/go-traceroute.svg)](https://travis-ci.org/pixelbender/go-traceroute)
[![Go Report Card](https://goreportcard.com/badge/github.com/pixelbender/go-traceroute)](https://goreportcard.com/report/github.com/pixelbender/go-traceroute)
[![GoDoc](https://godoc.org/github.com/pixelbender/go-traceroute?status.svg)](https://godoc.org/github.com/pixelbender/go-traceroute/traceroute)

## Features

- [x] Reusable and Concurrent raw IP socket (needs root)
- [x] IPv4 ID and IPv6 FlowLabel tracking
- [ ] ICMPv6 support
- [ ] UDP/TCP support
- [ ] Remove golang.org/x/net

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
    hops, err := traceroute.Trace(net.ParseIP("1.1.1.1"))
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
            Networks: []string{"ip4:icmp", "ip4:ip"},
        },
    }
    defer t.Close()
    err := t.Trace(context.Background(), net.ParseIP("1.1.1.1"), func(reply *traceroute.Reply) {
        log.Printf("%d. %v %v", reply.Hops, reply.IP, reply.RTT)
    })
    if err != nil {
        log.Fatal(err)
    }
}
```
