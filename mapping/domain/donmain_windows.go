//go:build windows
// +build windows

package domain

import "net"

type domain struct {
	conn *net.IPConn
}

func New() *domain {
	return nil
}

func newDomain() *domain {
	// conn, err := net.ListenIP("udp4", &net.IPAddr{})
	return nil
}

func (d *domain) service() (_ error) {
	return
}
