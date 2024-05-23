package pcap

import (
	"io"
	"net/netip"
	"os"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/lysShub/netkit/packet"
	"github.com/pkg/errors"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

type Pcap struct {
	mu sync.RWMutex
	fh io.Closer
	w  *pcapgo.Writer
}

func New(w io.Writer) (*Pcap, error) {
	return newPcap(w, false)
}

func File(file string) (*Pcap, error) {
	var exist bool = true

	fh, err := os.OpenFile(file, os.O_RDWR|os.O_APPEND, 0o666)
	if err != nil {
		if os.IsNotExist(err) {
			exist = false
			fh, err = os.Create(file)
		}

		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return newPcap(fh, exist)
}

func newPcap(w io.Writer, exist bool) (*Pcap, error) {
	p := pcapgo.NewWriter(w)
	if !exist {
		err := p.WriteFileHeader(0xffff, layers.LinkTypeEthernet)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	var pcap = &Pcap{w: p}
	if c, ok := w.(io.Closer); ok {
		pcap.fh = c
	}
	return pcap, nil
}

func (p *Pcap) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.fh != nil {
		return errors.WithStack(p.fh.Close())
	} else {
		return nil
	}
}

func (p *Pcap) write(eth header.Ethernet) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	err := p.w.WritePacket(gopacket.CaptureInfo{
		Timestamp:      time.Now(),
		CaptureLength:  len(eth),
		Length:         len(eth),
		InterfaceIndex: 0,
	}, eth)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (p *Pcap) Write(eth header.Ethernet) error {
	return p.write(eth)
}

func (p *Pcap) WriteIP(ip []byte) error {
	var eth []byte
	switch ver := header.IPVersion(ip); ver {
	case 4:
		eth = append(eth,
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x08, 0x00}...,
		)
	case 6:
		eth = append(eth,
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x86, 0xdd}...,
		)
	default:
		return errors.Errorf("not support ip version %d", ver)
	}
	eth = append(eth, ip...)

	return p.write(eth)
}

func (p *Pcap) WritePacket(ip *packet.Packet) error {
	defer ip.DetachN(header.EthernetMinimumSize)

	switch ver := header.IPVersion(ip.Bytes()); ver {
	case 4:
		ip.Attach([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x08, 0x00})
	case 6:
		ip.Attach([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x86, 0xdd})
	default:
		return errors.Errorf("not support ip version %d", ver)
	}
	return p.write(ip.Bytes())
}

func (p *Pcap) Overhead() (int, int) { return header.EthernetMinimumSize, 0 }

type BindPcap struct {
	*Pcap
	laddr netip.Addr
}

func Bind(p *Pcap, laddr netip.Addr) (*BindPcap, error) {
	if !laddr.Is4() {
		return nil, errors.New("only support ipv4")
	}
	return &BindPcap{
		Pcap:  p,
		laddr: laddr,
	}, nil
}

// Outbound pcap a outbound tcp/udp/icmp packet
func (b *BindPcap) Outbound(dst netip.Addr, proto tcpip.TransportProtocolNumber, p []byte) error {
	return b.write(b.laddr, dst, proto, p)
}

// Inbound pcap a inbound tcp/udp/icmp packet
func (b *BindPcap) Inbound(src netip.Addr, proto tcpip.TransportProtocolNumber, p []byte) error {
	return b.write(src, b.laddr, proto, p)
}

func (b *BindPcap) write(src, dst netip.Addr, proto tcpip.TransportProtocolNumber, p []byte) error {
	if !src.Is4() {
		return errors.New("only support ipv4")
	}
	ip := make(header.IPv4, len(p)+header.IPv4MinimumSize)
	copy(ip[header.IPv4MinimumSize:], p)

	ip.Encode(&header.IPv4Fields{
		TOS:            0,
		TotalLength:    uint16(len(ip)),
		ID:             0,
		Flags:          0,
		FragmentOffset: 0,
		TTL:            64,
		Protocol:       uint8(proto),
		Checksum:       0,
		SrcAddr:        tcpip.AddrFrom4(src.As4()),
		DstAddr:        tcpip.AddrFrom4(dst.As4()),
		Options:        nil,
	})
	ip.SetChecksum(^ip.CalculateChecksum())

	return b.Pcap.WriteIP(ip)
}

// WritePacket pcap a tcp/udp/icmp packet
func (b *BindPcap) WritePacket(src, dst netip.Addr, proto tcpip.TransportProtocolNumber, pkt *packet.Packet) error {
	if !src.Is4() {
		return errors.New("only support ipv4")
	}

	defer pkt.DetachN(header.IPv4MinimumSize)
	ip := header.IPv4(pkt.AttachN(header.IPv4MinimumSize).Bytes())
	ip.Encode(&header.IPv4Fields{
		TOS:            0,
		TotalLength:    uint16(len(ip)),
		ID:             0,
		Flags:          0,
		FragmentOffset: 0,
		TTL:            64,
		Protocol:       uint8(proto),
		Checksum:       0,
		SrcAddr:        tcpip.AddrFrom4(src.As4()),
		DstAddr:        tcpip.AddrFrom4(dst.As4()),
		Options:        nil,
	})
	ip.SetChecksum(^ip.CalculateChecksum())

	return b.Pcap.WritePacket(pkt)
}

func (p *BindPcap) Overhead() (int, int) { return packet.Inherit(p.Pcap, header.IPv4MinimumSize, 0) }
