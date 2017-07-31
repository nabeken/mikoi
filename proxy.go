package main

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

type ProxyConn struct {
	net.Conn

	once sync.Once
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
	saddr, sport, err := net.SplitHostPort(s.String())
	if err != nil {
		return err
	}

	d := c.Conn.RemoteAddr()
	daddr, dport, err := net.SplitHostPort(d.String())
	if err != nil {
		return err
	}

	raddr, ok := d.(*net.TCPAddr)
	if !ok {
		return errors.New("proxyconn: must be tcp or tcp6")
	}

	var tcpStr string
	if rip4 := raddr.IP.To4(); len(rip4) == net.IPv4len {
		tcpStr = "TCP4"
	} else if len(raddr.IP) == net.IPv6len {
		tcpStr = "TCP6"
	} else {
		return errors.New("proxyconn: unknown length")
	}

	_, err = fmt.Fprintf(c.Conn, "PROXY %s %s %s %s %s\r\n", tcpStr, saddr, daddr, sport, dport)
	return err
}
