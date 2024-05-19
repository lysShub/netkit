package route

import (
	"net/netip"
)

type Table []Entry

// Match match best route entry
func (t Table) Match(dst netip.Addr) Entry {
	return t.match(dst)
}

// InterfaceByAddr get interface index by address, return 0 means not found.
func (t Table) InterfaceByAddr(addr netip.Addr) (ifIdx uint32) {
	if addr.IsUnspecified() {
		return t[0].Interface
	}

	for _, e := range t {
		if e.Addr == addr {
			return e.Interface
		}
	}
	return 0
}

// AddrByInterface get interface addr by index.
func (t Table) AddrByInterface(ifIdx uint32) (addr netip.Addr) {
	for _, e := range t {
		if e.Interface == ifIdx {
			return e.Addr
		}
	}
	return netip.Addr{}
}

func (t Table) match(dst netip.Addr) Entry {
	for i := len(t) - 1; i >= 0; i-- {
		if t[i].Dest.Contains(dst) {
			return t[i]
		}
	}
	return Entry{}
}

// Loopback detect addr is loopback
func (t Table) Loopback(addr netip.Addr) bool {
	e := t.match(addr)
	return e.Valid() && e.Addr == addr && e.Dest.IsSingleIP()
}

func (t Table) String() string {
	var p = newPrinter()
	for _, e := range t {
		e.string(p)
	}
	return p.string()
}

type tableSortImpl Table

func (es tableSortImpl) Len() int { return len(es) }
func (es tableSortImpl) Less(i, j int) bool {
	bi, bj := es[i].Dest.Bits(), es[j].Dest.Bits()
	if bi <= bj {
		if bi == bj {
			return es[i].Metric <= es[j].Metric
		}
		return true
	}
	return false
}
func (es tableSortImpl) Swap(i, j int) { es[i], es[j] = es[j], es[i] }
