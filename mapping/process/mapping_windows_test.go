//go:build windows
// +build windows

package process

import (
	"errors"
	"net/netip"
	"testing"

	"github.com/lysShub/divert-go"
	"github.com/stretchr/testify/require"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

func TestXxxx(t *testing.T) {
	divert.MustLoad(divert.DLL)
	defer divert.Release()

	h, err := divert.Open("outbound and !loopback and ip and (tcp or udp)", divert.Network, 0, divert.ReadOnly|divert.Sniff)
	require.NoError(t, err)
	defer h.Close()

	prcos, err := New()
	require.NoError(t, err)

	var b = make([]byte, 1536)
	var notRecords int
	for i := 0; i < 0xff; i++ {
		n, err := h.Recv(b[:cap(b)], nil)
		require.NoError(t, err)

		laddr, proto := getLaddProto(t, b[:n])
		name, err := prcos.Name(laddr, proto)
		if err != nil {
			if errors.Is(err, ErrNotRecord{}) {
				notRecords++
				require.Less(t, notRecords, 5)
			} else {
				require.NoError(t, err)
			}
		}
		// fmt.Println(name)
		require.NotEqual(t, "", name)
	}
}

func getLaddProto(t *testing.T, ip header.IPv4) (netip.AddrPort, uint8) {
	addr := netip.AddrFrom4(ip.SourceAddress().As4())

	proto := ip.TransportProtocol()
	switch proto {
	case header.TCPProtocolNumber:
		tcp := header.TCP(ip.Payload())
		return netip.AddrPortFrom(addr, tcp.SourcePort()), uint8(proto)
	case header.UDPProtocolNumber:
		udp := header.UDP(ip.Payload())
		return netip.AddrPortFrom(addr, udp.SourcePort()), uint8(proto)
	default:
		require.Contains(t, []tcpip.TransportProtocolNumber{header.TCPProtocolNumber, header.UDPProtocolNumber}, proto)
		panic("")
	}
}
