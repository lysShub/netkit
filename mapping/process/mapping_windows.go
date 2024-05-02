//go:build windows
// +build windows

package process

import (
	"os"
	"slices"
	"sync/atomic"

	"github.com/pkg/errors"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

type mapping struct {
	tcp *table
	udp *table

	closeErr atomic.Pointer[error]
}

var _ Mapping = (*mapping)(nil)

func newMapping() (*mapping, error) {
	var m = &mapping{
		tcp: newTable(header.TCPProtocolNumber),
		udp: newTable(header.UDPProtocolNumber),
	}
	// m.query(Endpoint{Proto: header.UDPProtocolNumber})
	// m.query(Endpoint{Proto: header.TCPProtocolNumber})
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

func (m *mapping) query(ep ID) (elem, error) {
	switch ep.Proto {
	case header.TCPProtocolNumber:
		e := m.tcp.Query(ep.Local)
		if e.valid() {
			return e, nil
		}

		if err := m.tcp.Upgrade(); err != nil {
			return elem{}, err
		}

		e = m.tcp.Query(ep.Local)
		if e.valid() {
			return e, nil
		}
		return elem{}, nil // not record
	case header.UDPProtocolNumber:
		e := m.udp.Query(ep.Local)
		if e.valid() {
			return e, nil
		}

		if err := m.udp.Upgrade(); err != nil {
			return elem{}, err
		}

		e = m.udp.Query(ep.Local)
		if e.valid() {
			return e, nil
		}
		return elem{}, nil // not record
	default:
		return elem{}, errors.Errorf("not support protocol %s", protoStr(ep.Proto))
	}

}

func (m *mapping) Name(ep ID) (string, error) {
	if e, err := m.query(ep); err != nil {
		return "", err
	} else {
		return e.name, nil
	}
}

func (m *mapping) Pid(ep ID) (uint32, error) {
	if e, err := m.query(ep); err != nil {
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
