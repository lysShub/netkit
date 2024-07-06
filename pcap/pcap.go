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
		ip.Attach([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x08, 0x00}...)
	case 6:
		ip.Attach([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x86, 0xdd}...)
	default:
		return errors.Errorf("not support ip version %d", ver)
	}
	return p.write(ip.Bytes())
}

func (p *Pcap) WritePayload(src, dst netip.Addr, proto tcpip.TransportProtocolNumber, payload *packet.Packet) error {
	if src.Is4() && dst.Is4() {
		ip := header.IPv4(payload.AttachN(header.IPv4MinimumSize).Bytes())
		ip.Encode(&header.IPv4Fields{
			TotalLength: uint16(len(ip)),
			Protocol:    uint8(proto),
			SrcAddr:     tcpip.AddrFrom4(src.As4()),
			DstAddr:     tcpip.AddrFrom4(dst.As4()),
		})
	} else if src.Is6() && dst.Is6() {
		ip := header.IPv6(payload.AttachN(header.IPv6MinimumSize).Bytes())
		ip.Encode(&header.IPv6Fields{
			TrafficClass:  uint8(proto),
			PayloadLength: uint16(len(ip)) - header.IPv6MinimumSize,
			SrcAddr:       tcpip.AddrFrom16(src.As16()),
			DstAddr:       tcpip.AddrFrom16(dst.As16()),
		})
	} else {
		return errors.Errorf("src %s dst %s", src.String(), dst.String())
	}
	return nil
}
