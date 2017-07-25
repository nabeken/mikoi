package main

import (
    "net"
)

type Addr struct {
    IP net.IP
}

func (addr Addr) Network() string { return "tcp" }
func (addr Addr) String() string { return addr.IP.String() }
