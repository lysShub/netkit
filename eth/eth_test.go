package eth

import (
	"math/rand"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/checksum"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

func Test_Htons(t *testing.T) {
	a := Htons(header.IPv4ProtocolNumber)
	require.Equal(t, 8, int(a))
}

func Baidu() netip.Addr {
	ips, err := net.LookupIP("baidu.com")
	if err != nil {
		panic(err)
	}

	for _, e := range ips {
		if ip := e.To4(); ip != nil {
			return netip.AddrFrom4([4]byte(ip))
		}
	}
	panic("not found ")
}

func BuildICMP(t require.TestingT, src, dst netip.Addr, typ header.ICMPv4Type, msg []byte) header.IPv4 {
	require.Zero(t, len(msg)%4)

	var iphdr = make(header.IPv4, 28+len(msg))
	iphdr.Encode(&header.IPv4Fields{
		TOS:            0,
		TotalLength:    uint16(len(iphdr)),
		ID:             uint16(rand.Uint32()),
		Flags:          0,
		FragmentOffset: 0,
		TTL:            128,
		Protocol:       uint8(header.ICMPv4ProtocolNumber),
		Checksum:       0,
		SrcAddr:        tcpip.AddrFromSlice(src.AsSlice()),
		DstAddr:        tcpip.AddrFromSlice(dst.AsSlice()),
	})
	iphdr.SetChecksum(^checksum.Checksum(iphdr[:iphdr.HeaderLength()], 0))
	require.True(t, iphdr.IsChecksumValid())

	icmphdr := header.ICMPv4(iphdr.Payload())
	icmphdr.SetType(typ)
	icmphdr.SetCode(0)
	icmphdr.SetChecksum(0)
	icmphdr.SetIdent(0x0005)
	icmphdr.SetSequence(0x0001)
	copy(icmphdr.Payload(), msg)
	icmphdr.SetChecksum(^checksum.Checksum(icmphdr, 0))

	ValidIP(t, iphdr)
	return iphdr
}

func LocIP() netip.Addr {
	c, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
	if err != nil {
		panic(err)
	}
	defer c.Close()
	return netip.MustParseAddrPort(c.LocalAddr().String()).Addr()
}

func BuildRawTCP(t require.TestingT, laddr, raddr netip.AddrPort, payload []byte) header.IPv4 {
	require.True(t, laddr.Addr().Is4())

	iptcp := header.IPv4MinimumSize + header.TCPMinimumSize

	totalSize := iptcp + len(payload)
	var b = make([]byte, totalSize)
	copy(b[iptcp:], payload)

	ts := uint32(time.Now().UnixNano())
	tcphdr := header.TCP(b[header.IPv4MinimumSize:])
	tcphdr.Encode(&header.TCPFields{
		SrcPort:    uint16(laddr.Port()),
		DstPort:    uint16(raddr.Port()),
		SeqNum:     501 + ts,
		AckNum:     ts,
		DataOffset: header.TCPMinimumSize,
		Flags:      header.TCPFlagAck | header.TCPFlagPsh,
		WindowSize: 83,
		Checksum:   0,
	})

	// panic("todo")
	// s, err := ipstack.New(laddr.Addr(), raddr.Addr(), header.TCPProtocolNumber)
	// require.NoError(t, err)
	// p := packet.Make().Append(b).SetHead(s.Size())
	// s.AttachOutbound(p)

	// psoSum := s.AttachHeader(b, header.TCPProtocolNumber)

	// tcphdr.SetChecksum(^checksum.Checksum(tcphdr, psoSum))

	require.True(t, header.IPv4(b).IsChecksumValid())
	require.True(t,
		tcphdr.IsChecksumValid(
			tcpip.AddrFromSlice(laddr.Addr().AsSlice()),
			tcpip.AddrFromSlice(raddr.Addr().AsSlice()),
			checksum.Checksum(tcphdr.Payload(), 0),
			uint16(len(tcphdr.Payload())),
		),
	)

	return b
}
func ValidIP(t require.TestingT, ip []byte) {
	panic("todo")
}
