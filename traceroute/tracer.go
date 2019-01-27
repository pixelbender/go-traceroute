package traceroute

import (
	"context"
	"errors"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// DefaultConfig is the default configuration for Tracer.
var DefaultConfig = Config{
	Delay:   50 * time.Millisecond,
	Timeout: 2 * time.Second,
	MaxHops: 30,
	Count:   3,
	Network: "ip4:icmp",
}

// DefaultTracer is a tracer with DefaultConfig.
var DefaultTracer = &Tracer{
	Config: DefaultConfig,
}

// Config is a configuration for Tracer.
type Config struct {
	Delay   time.Duration
	Timeout time.Duration
	MaxHops int
	Count   int
	Network string
	Addr    *net.IPAddr
}

// Tracer is a traceroute tool based on raw IP packets.
// It can handle multiple sessions simultaneously.
type Tracer struct {
	Config

	once sync.Once
	conn *net.IPConn
	err  error

	mu   sync.RWMutex
	sess map[string][]*session
	seq  uint32
}

// Trace starts sending IP packets to ip with TTL = 1, 2, ..., MaxHops and calls h for each reply.
// It can be called concurrently.
func (t *Tracer) Trace(ctx context.Context, ip net.IP, h func(reply *Reply)) error {
	t.once.Do(t.init)
	if t.err != nil {
		return t.err
	}
	sess := newSession(t, ip)
	defer sess.Close()

	delay := time.NewTicker(t.Delay)
	m := make(map[uint16]*packet)
	handle := func(res *packet) bool {
		req, ok := m[res.ID]
		if !ok {
			return false
		}
		hops := req.TTL - res.TTL + 1
		if hops < 1 {
			hops = 1
		}
		h(&Reply{
			IP:   res.IP,
			RTT:  res.Time.Sub(req.Time),
			Hops: hops,
		})
		return true
	}
	for n := 0; n < t.Count; n++ {
		max := t.MaxHops
		for ttl := 1; ttl < max; ttl++ {
			req, err := sess.Probe(ttl)
			if err != nil {
				return err
			}
			m[req.ID] = req
		wait:
			for {
				select {
				case <-delay.C:
					break wait
				case res := <-sess.C:
					if handle(res) {
						if max > req.TTL && ip.Equal(res.IP) {
							max = req.TTL
						}
						break wait
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
	deadline := time.After(t.Timeout)
	for {
		select {
		case res := <-sess.C:
			handle(res)
		case <-deadline:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (t *Tracer) init() {
	t.conn, t.err = t.listen(t.Network, t.Addr)
	if t.err != nil {
		return
	}
	go t.serve(t.conn)
}

func (t *Tracer) listen(network string, laddr *net.IPAddr) (*net.IPConn, error) {
	conn, err := net.ListenIP(network, laddr)
	if err != nil {
		return nil, err
	}
	raw, err := conn.SyscallConn()
	if err != nil {
		conn.Close()
		return nil, err
	}
	_ = raw.Control(func(fd uintptr) {
		err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
	})
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// Close closes listening socket.
// Tracer can not be used after Close is called.
func (t *Tracer) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn != nil {
		t.conn.Close()
	}
}

func (t *Tracer) serve(conn *net.IPConn) error {
	defer conn.Close()
	buf := make([]byte, 1500)
	for {
		n, from, err := conn.ReadFromIP(buf)
		if err != nil {
			return err
		}
		err = t.serveData(from.IP, buf[:n])
		if err != nil {
			continue
		}
	}
}

func (t *Tracer) serveData(from net.IP, b []byte) error {
	if from.To4() == nil {
		// TODO: implement ProtocolIPv6ICMP
		return errUnsupportedProtocol
	}
	now := time.Now()
	msg, err := icmp.ParseMessage(ProtocolICMP, b)
	if err != nil {
		return err
	}
	if msg.Type == ipv4.ICMPTypeEchoReply {
		echo := msg.Body.(*icmp.Echo)
		return t.serveReply(from, &packet{from, uint16(echo.ID), 1, now})
	}
	b = getReplyData(msg)
	if len(b) < ipv4.HeaderLen {
		return errMessageTooShort
	}
	switch b[0] >> 4 {
	case ipv4.Version:
		ip, err := ipv4.ParseHeader(b)
		if err != nil {
			return err
		}
		return t.serveReply(ip.Dst, &packet{from, uint16(ip.ID), ip.TTL, now})
	case ipv6.Version:
		ip, err := ipv6.ParseHeader(b)
		if err != nil {
			return err
		}
		return t.serveReply(ip.Dst, &packet{from, uint16(ip.FlowLabel), ip.HopLimit, now})
	default:
		return errUnsupportedProtocol
	}
}

func (t *Tracer) serveReply(dst net.IP, res *packet) error {
	t.mu.RLock()
	a := t.sess[string(shortIP(dst))]
	t.mu.RUnlock()
	for _, it := range a {
		go it.Handle(res)
	}
	return nil
}

func (t *Tracer) sendRequest(dst net.IP, ttl int) (*packet, error) {
	id := uint16(atomic.AddUint32(&t.seq, 1))
	b := newPacket(id, dst, ttl)
	req := &packet{dst, id, ttl, time.Now()}
	_, err := t.conn.WriteToIP(b, &net.IPAddr{IP: dst})
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (t *Tracer) addSess(s *session) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sess == nil {
		t.sess = make(map[string][]*session)
	}
	t.sess[string(s.dst)] = append(t.sess[string(s.dst)], s)
}

func (t *Tracer) removeSess(s *session) {
	t.mu.Lock()
	defer t.mu.Unlock()
	a := t.sess[string(s.dst)]
	for i, it := range a {
		if it == s {
			t.sess[string(s.dst)] = append(a[:i], a[i+1:]...)
			return
		}
	}
}

type packet struct {
	IP   net.IP
	ID   uint16
	TTL  int
	Time time.Time
}

type session struct {
	*Tracer
	dst net.IP
	C   chan *packet
}

func newSession(t *Tracer, dst net.IP) *session {
	s := &session{
		t,
		shortIP(dst),
		make(chan *packet, 64),
	}
	t.addSess(s)
	return s
}

func (s *session) Handle(res *packet) {
	s.C <- res
}

func (s *session) Probe(ttl int) (*packet, error) {
	return s.sendRequest(s.dst, ttl)
}

func (s *session) Close() {
	s.removeSess(s)
}

func shortIP(ip net.IP) net.IP {
	if v := ip.To4(); v != nil {
		return v
	}
	return ip
}

func getReplyData(msg *icmp.Message) []byte {
	switch b := msg.Body.(type) {
	case *icmp.TimeExceeded:
		return b.Data
	case *icmp.DstUnreach:
		return b.Data
	case *icmp.ParamProb:
		return b.Data
	}
	return nil
}

var (
	errMessageTooShort     = errors.New("message too short")
	errUnsupportedProtocol = errors.New("unsupported protocol")
	errNoReplyData         = errors.New("no reply data")
)

func newPacket(id uint16, dst net.IP, ttl int) []byte {
	// TODO: reuse buffers...
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Body: &icmp.Echo{
			ID:  int(id),
			Seq: int(id),
		},
	}
	p, _ := msg.Marshal(nil)
	ip := &ipv4.Header{
		Version:  ipv4.Version,
		Len:      ipv4.HeaderLen,
		TotalLen: ipv4.HeaderLen + len(p),
		TOS:      16,
		ID:       int(id),
		Dst:      dst,
		Protocol: ProtocolICMP,
		TTL:      ttl,
	}
	buf, err := ip.Marshal()
	if err != nil {
		return nil
	}
	return append(buf, p...)
}

// IANA Assigned Internet Protocol Numbers
const (
	ProtocolICMP     = 1
	ProtocolTCP      = 6
	ProtocolUDP      = 17
	ProtocolIPv6ICMP = 58
)

// Reply is a reply packet.
type Reply struct {
	IP   net.IP
	RTT  time.Duration
	Hops int
}

// Node is a detected network node.
type Node struct {
	IP  net.IP
	RTT []time.Duration
}

// Hop is a set of detected nodes.
type Hop struct {
	Nodes    []*Node
	Distance int
}

// Add adds node from r.
func (h *Hop) Add(r *Reply) *Node {
	var node *Node
	for _, it := range h.Nodes {
		if it.IP.Equal(r.IP) {
			node = it
			break
		}
	}
	if node == nil {
		node = &Node{IP: r.IP}
		h.Nodes = append(h.Nodes, node)
	}
	node.RTT = append(node.RTT, r.RTT)
	return node
}

// Trace is a simple traceroute tool using DefaultTracer.
func Trace(ip net.IP) ([]*Hop, error) {
	hops := make([]*Hop, 0, DefaultTracer.MaxHops)
	touch := func(dist int) *Hop {
		for _, h := range hops {
			if h.Distance == dist {
				return h
			}
		}
		h := &Hop{Distance: dist}
		hops = append(hops, h)
		return h
	}
	err := DefaultTracer.Trace(context.Background(), ip, func(r *Reply) {
		touch(r.Hops).Add(r)
	})
	if err != nil && err != context.DeadlineExceeded {
		return nil, err
	}
	sort.Slice(hops, func(i, j int) bool {
		return hops[i].Distance < hops[j].Distance
	})
	last := len(hops) - 1
	for i := last; i >= 0; i-- {
		h := hops[i]
		if len(h.Nodes) == 1 && ip.Equal(h.Nodes[0].IP) {
			continue
		}
		if i == last {
			break
		}
		i++
		node := hops[i].Nodes[0]
		i++
		for _, it := range hops[i:] {
			node.RTT = append(node.RTT, it.Nodes[0].RTT...)
		}
		hops = hops[:i]
		break
	}
	return hops, nil
}
