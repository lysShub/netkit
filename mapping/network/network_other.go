//go:build !windows
// +build !windows

package network

type Network struct{}

func New() (n *Network, err error) {
	panic("not support")
}

func (n *Network) visit(proto uint8, fn func(es []Elem)) error { panic("not support") }
func (n *Network) Upgrade(proto uint8) error                   { panic("not support") }
