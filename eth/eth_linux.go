//go:build linux
// +build linux

package eth

import (
	"net"
	"os"
	"syscall"
	"time"

	"github.com/lysShub/netkit/errorx"
	netcall "github.com/lysShub/netkit/syscall"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

// https://man7.org/linux/man-pages/man7/packet.7.html
type ETHConn struct {
	proto tcpip.NetworkProtocolNumber
	ifi   *net.Interface
	fd    *os.File
	raw   syscall.RawConn
}

var _ net.Conn = (*ETHConn)(nil)

func Listen(network string, ifi *net.Interface) (*ETHConn, error) {
	var proto tcpip.NetworkProtocolNumber
	switch network {
	case "eth:ip", "eth:ip4":
		proto = header.IPv4ProtocolNumber // unix.ETH_P_IP
	case "eth:ip6":
		proto = header.IPv6ProtocolNumber // unix.ETH_P_IPV6
	default:
		// todo: support unix.ETH_P_ALL
		return nil, errors.Errorf("not support network %s", network)
	}

	fd, err := unix.Socket(unix.AF_PACKET, unix.SOCK_DGRAM, int(netcall.Hton(uint16(proto))))
	if err != nil {
		return nil, err
	}

	if err = unix.Bind(fd, &unix.SockaddrLinklayer{
		Protocol: netcall.Hton(uint16(proto)),
		Ifindex:  ifi.Index,
		Pkttype:  unix.PACKET_HOST,
	}); err != nil {
		return nil, err
	}

	// for support deadline
	if err = unix.SetNonblock(fd, true); err != nil {
		return nil, err
	}

	f := os.NewFile(uintptr(fd), "")
	raw, err := f.SyscallConn()
	if err != nil {
		unix.Close(fd)
		return nil, err
	}

	return &ETHConn{
		proto: proto,
		ifi:   ifi,
		fd:    f,
		raw:   raw,
	}, nil
}

// todo: support dial

func (c *ETHConn) Read(eth []byte) (n int, err error) {
	n, from, err := c.ReadFromETH(eth[header.EthernetMinimumSize:])
	if err != nil {
		return 0, err
	}

	header.Ethernet(eth).Encode(&header.EthernetFields{
		SrcAddr: tcpip.LinkAddress(from),
		DstAddr: tcpip.LinkAddress(c.ifi.HardwareAddr),
		Type:    c.proto,
	})
	return n + header.EthernetMinimumSize, nil
}

func (c *ETHConn) ReadFromETH(ip []byte) (n int, from net.HardwareAddr, err error) {
	var src unix.Sockaddr
	var operr error
	if err = c.raw.Read(func(fd uintptr) (done bool) {
		// sometime, it return n is greater 6 than actual size,
		_, src, operr = unix.Recvfrom(int(fd), ip, unix.MSG_TRUNC)
		return opdone(operr)
	}); err != nil {
		return 0, nil, err
	}
	if operr != nil {
		return 0, nil, operr
	}

	switch header.IPVersion(ip) {
	case 4:
		n = int(header.IPv4(ip).TotalLength())
	case 6:
		n = int(header.IPv6(ip).PayloadLength() + header.IPv6FixedHeaderSize)
	default:
		return 0, nil, errors.Errorf("recved invalid ip packet: %#v", ip[:min(20, len(ip))])
	}
	if n > len(ip) {
		return 0, nil, errorx.ShortBuff(n, len(ip))
	}

	if src, ok := src.(*unix.SockaddrLinklayer); ok {
		from = src.Addr[:src.Halen]
	}
	return n, from, nil
}

func (c *ETHConn) Write(eth []byte) (n int, err error) {
	to := net.HardwareAddr(header.Ethernet(eth).DestinationAddress())
	n, err = c.WriteToETH(eth[header.EthernetMinimumSize:], to)
	if err != nil {
		return 0, err
	}
	return n + header.EthernetMinimumSize, nil
}

func (c *ETHConn) WriteToETH(ip []byte, hw net.HardwareAddr) (int, error) {
	dst := &unix.SockaddrLinklayer{
		Protocol: netcall.Hton(uint16(c.proto)),
		Ifindex:  c.ifi.Index,
		Pkttype:  unix.PACKET_HOST,
		Halen:    uint8(len(hw)),
	}
	copy(dst.Addr[:], hw)

	var err, operr error
	if err = c.raw.Write(func(fd uintptr) (done bool) {
		operr = unix.Sendto(int(fd), ip, 0, dst)
		return opdone(operr)
	}); err != nil {
		return 0, err
	}
	if operr != nil {
		return 0, err
	}

	return len(ip), nil
}

func (c *ETHConn) Close() error                          { return c.fd.Close() }
func (c *ETHConn) LocalAddr() net.Addr                   { return ETHAddr(c.ifi.HardwareAddr) }
func (c *ETHConn) RemoteAddr() net.Addr                  { return nil }
func (c *ETHConn) SyscallConn() (syscall.RawConn, error) { return c.raw, nil }
func (c *ETHConn) SetDeadline(t time.Time) error         { return c.fd.SetDeadline(t) }
func (c *ETHConn) SetReadDeadline(t time.Time) error     { return c.fd.SetReadDeadline(t) }
func (c *ETHConn) SetWriteDeadline(t time.Time) error    { return c.fd.SetWriteDeadline(t) }
func (c *ETHConn) Interface() *net.Interface             { return c.ifi }

type ETHAddr net.HardwareAddr

func (e ETHAddr) Network() string { return "eth" }
func (e ETHAddr) String() string  { return net.HardwareAddr(e).String() }

func opdone(operr error) bool {
	return operr != syscall.EWOULDBLOCK && operr != syscall.EAGAIN
}
