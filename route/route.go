package route

import (
	"net/netip"
	"slices"
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
	slices.SortStableFunc(t, func(i, j Entry) int {
		return -i.Less(j) // 倒序
	})
}
