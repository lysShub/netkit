package tun_test

import (
	"math/rand"
	"net"
	"net/netip"
)

func LocIP() netip.Addr {
	c, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
	if err != nil {
		panic(err)
	}
	defer c.Close()
	return netip.MustParseAddrPort(c.LocalAddr().String()).Addr()
}
func RandPort() uint16 {
	p := uint16(rand.Uint32())
	if p < 1024 {
		p += 1536
	} else if p > 0xffff-64 {
		p -= 128
	}
	return p
}
