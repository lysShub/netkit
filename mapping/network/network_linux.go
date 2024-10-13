//go:build linux
// +build linux

package network

type network struct {
}

func New() (*network, error) {
	panic("unimplement")
}

func (n *network) Upgrade(proto uint8) error                          { panic("not support") }
func (n *network) Query(proto uint8, fn func(Elem) (stop bool)) error { panic("not support") }
func (n *network) visit(proto uint8, fn func(es []Elem)) error        { panic("not support") }
