package process

import (
	"cmp"
	"net/netip"
	"slices"
	"sort"
	"sync"

	"github.com/pkg/errors"
	pnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type table struct {
	mu    sync.RWMutex
	proto uint8
	elems []elem // asc by port
}

type elem struct {
	port uint16
	pid  uint32
	addr netip.Addr
	name string
}

func (e elem) valid() bool { return e.pid != 0 }

func newTable(proto uint8) *table {
	return &table{
		proto: proto,
		elems: make([]elem, 0, 16),
	}
}

func (t *table) Query(addr netip.AddrPort) elem {
	t.mu.RLock()
	defer t.mu.RUnlock()

	port := addr.Port()
	n := len(t.elems)
	i := sort.Search(n, func(i int) bool {
		return t.elems[i].port >= port
	})
	for ; i < n; i++ {
		if t.elems[i].port == port {
			if t.elems[i].addr == addr.Addr() || t.elems[i].addr.IsUnspecified() {
				return t.elems[i]
			}
		} else {
			break
		}
	}
	return elem{}
}

func (t *table) Upgrade() error {
	ss, err := pnet.Connections(protoStr(t.proto))
	if err != nil {
		return errors.WithStack(err)
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.elems = t.elems[:0]

	for _, e := range ss {
		name, err := (&process.Process{Pid: e.Pid}).Name()
		if err != nil {
			continue
		}
		t.elems = append(t.elems, elem{
			port: uint16(e.Laddr.Port),
			pid:  uint32(e.Pid),
			addr: netip.MustParseAddr(e.Laddr.IP),
			name: name,
		})
	}

	slices.SortFunc(t.elems, func(a, b elem) int {
		return cmp.Compare(a.port, b.port)
	})

	return nil
}

func (t *table) Pids(pids []uint32) []uint32 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, e := range t.elems {
		pids = append(pids, e.pid)
	}
	return pids
}
func (t *table) Names(names []string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, e := range t.elems {
		names = append(names, e.name)
	}
	return names
}
func (t *table) size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.elems)
}
