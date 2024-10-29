package domain

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/pkg/errors"
)

type TcpAssembler struct {
	assembler *tcpassembly.Assembler

	mu    sync.RWMutex
	flows map[uint64]*stream
}

func NewTcpAssembler() *TcpAssembler {
	var a = &TcpAssembler{
		flows: map[uint64]*stream{},
	}
	a.assembler = tcpassembly.NewAssembler(tcpassembly.NewStreamPool((*StreamFactory)(a)))
	return a
}

func (a *TcpAssembler) Put(ip []byte, stamp time.Time) (data []byte, err error) {
	defer a.garbage()
	pkg := gopacket.NewPacket(ip, layers.LayerTypeIPv4, gopacket.Default)

	layer := pkg.TransportLayer()
	if tcp, ok := layer.(*layers.TCP); ok {
		netflow := pkg.NetworkLayer().NetworkFlow()
		a.assembler.AssembleWithTimestamp(netflow, tcp, stamp)
		hash := netflow.FastHash() ^ pkg.TransportLayer().TransportFlow().FastHash()

		a.mu.RLock()
		f, has := a.flows[hash]
		a.mu.RUnlock()
		if has && f.complete {
			data = f.data
			a.mu.Lock()
			delete(a.flows, hash)
			a.mu.Unlock()
			return data, nil
		}
		return nil, nil
	} else {
		return nil, errors.Errorf("not tcp packet %T", layer)
	}
}

func (a *TcpAssembler) garbage() {
	if time.Now().Second() == 0 {
		a.mu.Lock()
		defer a.mu.Unlock()

		for hash, flow := range a.flows {
			if time.Since(flow.update) > time.Minute {
				delete(a.flows, hash)
			}
		}
	}
}

func (a *TcpAssembler) newflow(netFlow, tcpFlow gopacket.Flow) tcpassembly.Stream {
	hash := netFlow.FastHash() ^ tcpFlow.FastHash()

	stream := &stream{data: make([]byte, 0, 256)}
	a.mu.Lock()
	a.flows[hash] = stream
	a.mu.Unlock()
	return stream
}

type StreamFactory TcpAssembler

func (sf *StreamFactory) New(netFlow, tcpFlow gopacket.Flow) tcpassembly.Stream {
	return (*TcpAssembler)(sf).newflow(netFlow, tcpFlow)
}

type stream struct {
	data     []byte
	update   time.Time
	complete bool
}

func (s *stream) Reassembled(reassembly []tcpassembly.Reassembly) {
	for _, reasm := range reassembly {
		s.data = append(s.data, reasm.Bytes...)
	}
	s.update = time.Now()
}

func (s *stream) ReassemblyComplete() { s.complete = true }

// https://www.rfc-editor.org/rfc/rfc7766#section-8
type RawDnsOverTcp []byte

func (d RawDnsOverTcp) Msgs() (msgs [][]byte) {
	if len(d) <= 2 {
		return nil
	}
	for i := 0; i < len(d); {
		n := int(binary.BigEndian.Uint16(d[i:]))
		if i+2+n <= len(d) {
			msgs = append(msgs, d[i+2:i+2+n])
			i = i + 2 + n
		} else {
			break // damaged!!
		}
	}
	return msgs
}
