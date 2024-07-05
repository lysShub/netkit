package syscall_test

import (
	"testing"

	"github.com/lysShub/netkit/syscall"
	"github.com/stretchr/testify/require"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

func Test_ReserveByte(t *testing.T) {
	var a uint16 = 0x0102
	b := syscall.ReserveByte(a)
	require.Equal(t, uint16(0x0201), b)
}

func Test_Hton(t *testing.T) {
	a := syscall.Hton(uint16(header.IPv4ProtocolNumber))
	require.Equal(t, 8, int(a))
}
