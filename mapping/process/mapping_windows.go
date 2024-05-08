//go:build windows
// +build windows

package process

import (
	"fmt"
	"net/netip"
	"os"
	"slices"
	"sync/atomic"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

type mapping struct {
	tcp *table
	udp *table

	closeErr atomic.Pointer[error]
}

var _ Mapping = (*mapping)(nil)

func newMapping() (*mapping, error) {
	var m = &mapping{
		tcp: newTable(windows.IPPROTO_TCP),
		udp: newTable(windows.IPPROTO_UDP),
	}
	// m.query(Endpoint{Proto: windows.IPPROTO_UDP})
	// m.query(Endpoint{Proto: windows.IPPROTO_TCP})
	return m, nil
}

func (m *mapping) close(cause error) error {
	if m.closeErr.CompareAndSwap(nil, &os.ErrClosed) {

		if cause != nil {
			m.closeErr.Store(&cause)
		}
		return cause
	}
	return *m.closeErr.Load()
}

func (m *mapping) query(laddr netip.AddrPort, proto uint8) (elem, error) {
	switch proto {
	case windows.IPPROTO_TCP:
		e := m.tcp.Query(laddr)
		if e.valid() {
			return e, nil
		}

		if err := m.tcp.Upgrade(); err != nil {
			return elem{}, err
		}

		e = m.tcp.Query(laddr)
		if e.valid() {
			return e, nil
		}
		return elem{}, ErrNotRecord{}
	case windows.IPPROTO_UDP:
		e := m.udp.Query(laddr)
		if e.valid() {
			return e, nil
		}

		if err := m.udp.Upgrade(); err != nil {
			return elem{}, err
		}

		e = m.udp.Query(laddr)
		if e.valid() {
			return e, nil
		}
		return elem{}, ErrNotRecord{}
	default:
		return elem{}, errors.Errorf("not support protocol %s", protoStr(proto))
	}

}

func (m *mapping) Name(laddr netip.AddrPort, proto uint8) (string, error) {
	if e, err := m.query(laddr, proto); err != nil {
		return "", err
	} else {
		return e.name, nil
	}
}

func (m *mapping) Pid(laddr netip.AddrPort, proto uint8) (uint32, error) {
	if e, err := m.query(laddr, proto); err != nil {
		return 0, err
	} else {
		return e.pid, nil
	}
}

func (m *mapping) Pids() []uint32 {
	var pids = make([]uint32, 0, m.tcp.size()+m.udp.size())
	pids = m.tcp.Pids(pids)
	pids = m.udp.Pids(pids)
	slices.Sort(pids)
	return slices.Compact(pids)
}

func (m *mapping) Names() []string {
	var names = make([]string, 0, m.tcp.size()+m.udp.size())
	names = m.tcp.Names(names)
	names = m.udp.Names(names)
	slices.Sort(names)
	return slices.Compact(names)
}
func (m *mapping) Close() error { return m.close(nil) }

func protoStr(proto uint8) string {
	switch proto {
	case windows.IPPROTO_TCP:
		return "tcp"
	case windows.IPPROTO_UDP:
		return "udp"
	case windows.IPPROTO_ICMP:
		return "icmp"
	case windows.IPPROTO_ICMPV6:
		return "icmp6"
	default:
		return fmt.Sprintf("unknown(%d)", int(proto))
	}
}
