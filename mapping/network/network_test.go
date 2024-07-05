package network

import (
	"net/netip"
	"syscall"
	"testing"
)

func Benchmark(b *testing.B) {
	laddr := netip.MustParseAddrPort("0.0.0.0:55555")
	Process(laddr, syscall.IPPROTO_TCP)

	for i := 0; i < b.N; i++ {
		Process(laddr, syscall.IPPROTO_TCP)
	}
}
