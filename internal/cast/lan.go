package cast

import (
	"context"
	"net"
	"time"
)

// packet is a datagram received in answer to a query.
type packet struct {
	data []byte
	src  *net.UDPAddr
}

// query sends msg to dst now and every interval until ctx ends, calling
// recv with each answer, one at a time, on the calling goroutine.
//
// A multicast query goes out on every network adapter: an unbound socket
// sends it on the one Windows picks, which is often a virtual adapter
// (Hyper-V, WSL, a VPN), and devices on the real network never see it.
func query(ctx context.Context, dst *net.UDPAddr, msg []byte, interval time.Duration, recv func(packet)) error {
	conns, err := openQueryConns(dst)
	if err != nil {
		return err
	}
	in := make(chan packet)
	for _, c := range conns {
		defer c.Close()
		go func() {
			for {
				buf := make([]byte, 9000)
				n, src, err := c.ReadFromUDP(buf)
				if err != nil {
					return
				}
				select {
				case in <- packet{buf[:n], src}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	send := func() {
		for _, c := range conns {
			c.WriteToUDP(msg, dst)
		}
	}
	send()
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			send()
		case p := <-in:
			recv(p)
		}
	}
}

// openQueryConns opens a socket on each of this computer's IPv4
// addresses for a multicast dst, or one unbound socket otherwise.
func openQueryConns(dst *net.UDPAddr) ([]*net.UDPConn, error) {
	var conns []*net.UDPConn
	if dst.IP.IsMulticast() {
		for _, ip := range multicastAddrs() {
			if c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: ip}); err == nil {
				conns = append(conns, c)
			}
		}
	}
	if len(conns) == 0 {
		c, err := net.ListenUDP("udp4", &net.UDPAddr{})
		if err != nil {
			return nil, err
		}
		conns = append(conns, c)
	}
	return conns, nil
}

// multicastAddrs lists the IPv4 addresses of the adapters that are up
// and can multicast, loopback left out.
func multicastAddrs() []net.IP {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []net.IP
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagMulticast == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := ifc.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			if n, ok := a.(*net.IPNet); ok && n.IP.To4() != nil && !n.IP.IsLinkLocalUnicast() {
				out = append(out, n.IP.To4())
			}
		}
	}
	return out
}
