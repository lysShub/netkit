package network

/*
	mapping of netowrk <==> process
*/

import (
	"net/netip"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/lysShub/netkit/mapping"
	"github.com/pkg/errors"
)

const (
	SystemName        = "system"
	SystemIdleName    = "system-idle"
	SystemProcessName = "system-proc"
)

type Elem struct {
	Laddr netip.AddrPort
	Raddr netip.AddrPort
	Proto uint8
	Pid   uint32
	Path  string
}

var (
	once   sync.Once
	global *Network
)

func Process(laddr netip.AddrPort, proto uint8) (e Elem, err error) {
	if !laddr.Addr().Is4() {
		return Elem{}, errors.Errorf("only support ipv4 %s", laddr.Addr().String())
	}
	n, err := New()
	if err != nil {
		return Elem{}, err
	}

	if err = n.visit(proto, func(es []Elem) {
		i := slices.IndexFunc(es, func(e Elem) bool { return equal(e.Laddr, laddr) })
		if i > 0 {
			e = es[i]
		}
	}); err != nil {
		return Elem{}, err
	}

	if e == (Elem{}) {
		if err := n.Upgrade(proto); err != nil {
			return Elem{}, err
		}
	}

	if err = n.visit(proto, func(es []Elem) {
		i := slices.IndexFunc(es, func(e Elem) bool { return equal(e.Laddr, laddr) })
		if i >= 0 {
			e = es[i]
		}
	}); err != nil {
		return Elem{}, err
	}
	if e == (Elem{}) {
		return Elem{}, mapping.ErrNotRecord{}
	}
	return e, nil
}

func AsyncProcess(laddr netip.AddrPort, proto uint8) (e Elem, err error) {
	if !laddr.Addr().Is4() {
		return Elem{}, errors.Errorf("only support ipv4 %s", laddr.Addr().String())
	}
	n, err := New()
	if err != nil {
		return Elem{}, err
	}

	if err = n.visit(proto, func(es []Elem) {
		i := slices.IndexFunc(es, func(e Elem) bool { return equal(e.Laddr, laddr) })
		if i >= 0 {
			e = es[i]
		}
	}); err != nil {
		return Elem{}, err
	}

	if e == (Elem{}) {
		go func() { n.Upgrade(proto) }()
		return Elem{}, mapping.ErrNotRecord{}
	} else {
		return e, nil
	}
}

func equal(a, b netip.AddrPort) bool {
	return a.Port() == b.Port() &&
		(a.Addr() == b.Addr() || a.Addr().IsUnspecified() || b.Addr().IsUnspecified())
}

func List(process string) (es []Elem, err error) {
	if process == "" {
		return es, err
	}
	n, err := New()
	if err != nil {
		return nil, err
	}

	if err := n.Upgrade(0); err != nil {
		return nil, err
	}

	if err := n.visit(syscall.IPPROTO_TCP, func(es1 []Elem) {
		for _, e := range es1 {
			if filepath.Base(e.Path) == process {
				es = append(es1, e)
			}
		}
	}); err != nil {
		return nil, err
	}
	if err := n.visit(syscall.IPPROTO_UDP, func(es1 []Elem) {
		for _, e := range es1 {
			if filepath.Base(e.Path) == process {
				es = append(es1, e)
			}
		}
	}); err != nil {
		return nil, err
	}
	return es, nil
}

func String(proto uint8) (string, error) {
	// laddr raddr(opt) proto name pid
	const (
		addrwidth = 24
		namewidth = 24
		pidwidth  = 6
		space     = 4
	)

	var s = &strings.Builder{}

	n, err := New()
	if err != nil {
		return "", err
	}

	if err := n.Upgrade(proto); err != nil {
		return "", err
	}

	var (
		protos []uint8
		pstrs  = map[uint8]string{
			syscall.IPPROTO_TCP: "tcp",
			syscall.IPPROTO_UDP: "udp",
		}
	)
	switch proto {
	case syscall.IPPROTO_TCP, syscall.IPPROTO_UDP:
		protos = []uint8{proto}
	case 0:
		protos = []uint8{syscall.IPPROTO_TCP, syscall.IPPROTO_UDP}
	default:
		return "", errors.Errorf("not support protocol %d", proto)
	}

	for _, p := range protos {
		n.visit(p, func(es []Elem) {
			for _, e := range es {
				fill(s, e.Laddr.String(), addrwidth)
				fill(s, "", space)

				if e.Raddr.IsValid() {
					fill(s, e.Raddr.String(), addrwidth)
				} else {
					fill(s, "", addrwidth)
				}
				fill(s, "", space)

				fill(s, pstrs[p], 3)
				fill(s, "", space)

				fill(s, filepath.Base(e.Path), namewidth)
				fill(s, "", space)

				fill(s, strconv.Itoa(int(e.Pid)), pidwidth)
				fill(s, "", space)

				s.WriteRune('\n')
			}
		})
	}
	return s.String(), nil
}

func fill(s *strings.Builder, str string, size int) {
	s.WriteString(str[:min(len(str), size)])
	for i := 0; i < size-len(str); i++ {
		s.WriteRune(' ')
	}
}
