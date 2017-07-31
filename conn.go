package main

import "net"

// mikoiConn implements net.Conn interface just for spoofing src addr in the proxyproto header.
type mikoiConn struct {
	net.Conn

	Src net.Addr
}

func (c *mikoiConn) LocalAddr() net.Addr {
	return c.Src
}
