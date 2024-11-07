package route

import (
	"net/netip"
	"slices"
)

type Table []Entry

func (t Table) MatchFunc(dst netip.Addr, fn func(Entry) (hit bool)) Entry {
	return t.matchFunc(dst, fn)
}

// Match match best route entry
func (t Table) Match(dst netip.Addr) Entry {
	return t.matchFunc(dst, nil)
}

func (t Table) matchFunc(dst netip.Addr, fn func(Entry) (hit bool)) Entry {
	type info struct {
		i         int
		prefixLen uint8
		metrics   uint32
	}
	var infos []info
	for i := len(t) - 1; i >= 0; i-- {
		if t[i].Addr.IsValid() && t[i].Dest.Contains(dst) {
			if fn == nil || fn(t[i]) {
				infos = append(infos, info{
					i:         i,
					prefixLen: uint8(t[i].Dest.Bits()),
					metrics:   t[i].Metric,
				})
			}
		}
	}
	if len(infos) == 0 {
		return Entry{}
	}

	slices.SortFunc(infos, func(a, b info) int {
		if a.prefixLen < b.prefixLen {
			return -1
		} else if a.prefixLen > b.prefixLen {
			return 1
		} else {
			if a.metrics < b.metrics {
				return 1
			} else if a.metrics > b.metrics {
				return -1
			} else {
				return 0
			}
		}
	})
	return t[infos[len(infos)-1].i]
}

// Loopback detect addr is loopback
func (t Table) Loopback(addr netip.Addr) bool {
	e := t.matchFunc(addr, nil)
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
