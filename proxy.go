package main

import (
	"errors"
	"net"
	"strconv"
	"sync"

	"github.com/nabeken/go-proxyproto"
)

type ProxyConn struct {
	net.Conn

	version int
	once    sync.Once
}

func (c *ProxyConn) Read(b []byte) (int, error) {
	var err error
	c.once.Do(func() { err = c.writeProxyProtocolHeader() })
	if err != nil {
		return 0, err
	}
	return c.Conn.Read(b)
}

func (c *ProxyConn) Write(b []byte) (int, error) {
	var err error
	c.once.Do(func() { err = c.writeProxyProtocolHeader() })
	if err != nil {
		return 0, err
	}
	return c.Conn.Write(b)
}

func (c *ProxyConn) writeProxyProtocolHeader() error {
	s := c.Conn.LocalAddr()
	saddr, sportStr, err := net.SplitHostPort(s.String())
	if err != nil {
		return err
	}

	d := c.Conn.RemoteAddr()
	daddr, dportStr, err := net.SplitHostPort(d.String())
	if err != nil {
		return err
	}

	raddr, ok := d.(*net.TCPAddr)
	if !ok {
		return errors.New("proxyconn: must be tcp or tcp6")
	}

	sport, _ := strconv.Atoi(sportStr)
	dport, _ := strconv.Atoi(dportStr)

	hdr := &proxyproto.Header{
		Version: c.version,
		SrcAddr: net.ParseIP(saddr),
		DstAddr: net.ParseIP(daddr),
		SrcPort: uint16(sport),
		DstPort: uint16(dport),

		Command: proxyproto.PROXY,
	}

	if rip4 := raddr.IP.To4(); len(rip4) == net.IPv4len {
		hdr.TransportProtocol = proxyproto.TCPv4
	} else if len(raddr.IP) == net.IPv6len {
		hdr.TransportProtocol = proxyproto.TCPv6
	} else {
		return errors.New("proxyconn: unknown length")
	}

	_, err = hdr.WriteTo(c.Conn)
	return err
}
