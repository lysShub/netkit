package route

import (
	"net/netip"
	"sort"
)

type Table []Entry

// Match match best route entry
func (t Table) Match(dst netip.Addr) Entry {
	return t.matchFunc(dst, nil)
}

func (t Table) MatchFunc(dst netip.Addr, fn func(Entry) (hit bool)) Entry {
	return t.matchFunc(dst, fn)
}

func (t Table) matchFunc(dst netip.Addr, fn func(Entry) (hit bool)) Entry {
	for i := len(t) - 1; i >= 0; i-- {
		if t[i].Addr.IsValid() && t[i].Dest.Contains(dst) {
			if fn == nil || fn(t[i]) {
				return t[i]
			}
		}
	}
	return Entry{}
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

func (t Table) Sort() {
	sort.Sort(tableSortImpl(t))
}

type tableSortImpl Table

func (es tableSortImpl) Len() int { return len(es) }
func (es tableSortImpl) Less(i, j int) bool {
	bi, bj := es[i].Dest.Bits(), es[j].Dest.Bits()
	if bi <= bj {
		if bi == bj {
			if es[i].Metric == es[j].Metric {
				return es[i].Dest.Addr().Less(es[j].Dest.Addr())
			}
			return es[i].Metric > es[j].Metric
		}
		return true
	}
	return false
}
func (es tableSortImpl) Swap(i, j int) { es[i], es[j] = es[j], es[i] }
