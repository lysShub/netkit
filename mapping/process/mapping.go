package process

import (
	"fmt"
	"net/netip"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

// mapping of local-addr <==> process
type Mapping interface {
	Close() error

	Name(ID) (string, error)
	Pid(ID) (uint32, error)

	Pids() []uint32
	Names() []string
}

type LocalAddr struct {
	Proto tcpip.TransportProtocolNumber
	Addr  netip.AddrPort
}

func New() (Mapping, error) {
	return newMapping()
}

type ID struct {
	Local netip.AddrPort                // local address
	Proto tcpip.TransportProtocolNumber // protocol
}

func (e ID) Valid() bool {
	return e.Proto != 0 && e.Local.IsValid()
}

func protoStr(num tcpip.TransportProtocolNumber) string {
	switch num {
	case header.TCPProtocolNumber:
		return "tcp"
	case header.UDPProtocolNumber:
		return "udp"
	case header.ICMPv4ProtocolNumber:
		return "icmp"
	case header.ICMPv6ProtocolNumber:
		return "icmp6"
	default:
		return fmt.Sprintf("unknown(%d)", int(num))
	}
}
