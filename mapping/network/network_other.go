//go:build !windows && !linux
// +build !windows,!linux

package network

type network struct{}

func New() (n *network, err error) {
	panic("not support")
}

func (n *network) visit(proto uint8, fn func(es []Elem)) error { panic("not support") }
func (n *network) Upgrade(proto uint8) error                   { panic("not support") }
