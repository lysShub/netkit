//go:build linux
// +build linux

package network

type network struct {
}

func New() (*network, error) {
	panic("unimplement")
}

func (n *network) Upgrade(proto uint8) error
func (n *network) Query(proto uint8, fn func(Elem) (stop bool)) error
func (n *network) visit(proto uint8, fn func(es []Elem)) error
