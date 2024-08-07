package packet

import (
	"log/slog"
	"math"

	"github.com/lysShub/netkit/debug"
	"github.com/lysShub/netkit/errorx"
)

type Packet struct {
	//             |    data     |
	//  b  --------+++++++++++++++--------
	//     |  head |             | tail  |
	//
	// 	head = i
	// 	tail = cap(b) - len(b)

	i int
	b []byte
}

func From(b []byte) *Packet {
	return &Packet{b: b}
}

func Make(ns ...int) *Packet {
	var (
		head int = DefaulfHead
		n    int = 0
		tail int = DefaulfTail
	)
	if len(ns) > 0 {
		head = ns[0]
	}
	if len(ns) > 1 {
		n = ns[1]
	}
	if len(ns) > 2 {
		tail = ns[2]
	}

	return &Packet{
		i: head,
		b: make([]byte, head+n, head+n+tail),
	}
}

const (
	DefaulfHead = 32
	DefaulfTail = 16
)

func (p *Packet) Bytes() []byte {
	return p.b[p.i:]
}

// Head head section size
func (p *Packet) Head() int { return p.i }

// Data data section size
func (p *Packet) Data() int { return len(p.b) - p.i }

// Tail tail section size
func (p *Packet) Tail() int { return cap(p.b) - len(p.b) }

func (p *Packet) SetHead(head int) *Packet {
	p.i = min(max(head, 0), len(p.b))
	return p
}

func (p *Packet) SetData(data int) *Packet {
	if debug.Debug() && data == math.MaxInt {
		slog.Warn("overflow warning", errorx.Trace(nil))
	}

	p.b = p.b[:min(p.Head()+max(data, 0), cap(p.b))]
	return p
}

// Sets set head and data section size, equivalent to:
func (p *Packet) Sets(head, data int) *Packet {
	p.SetHead(head)
	return p.SetData(data)
}

// Attach attach b ahead data-section, use head-section firstly, if head section too short,
// will re-alloc memory.
func (p *Packet) Attach(b ...byte) *Packet {
	copy(p.AttachN(len(b)).Bytes(), b)
	return p
}

func (p *Packet) AttachN(n int) *Packet {
	head := p.Head() - max(n, 0)
	if head >= 0 {
		p.i = head
	} else {
		if debug.Debug() {
			slog.Warn("packet memory alloc", errorx.Trace(nil))
		}

		size := len(p.b) - head + DefaulfHead
		tmp := make([]byte, size, size+p.Tail())
		copy(tmp[DefaulfHead+n:], p.Bytes())

		p.b = tmp
		p.i = DefaulfHead
	}
	return p
}

func (p *Packet) Detach(n int) []byte {
	b := p.Bytes()
	p.DetachN(n)
	return b[:min(n, len(b))]
}

func (p *Packet) DetachTo(to []byte) []byte {
	n := copy(to, p.Detach(len(to)))
	return to[:n]
}

func (p *Packet) DetachN(n int) *Packet {
	p.i += min(max(n, 0), p.Data())
	return p
}

func (p *Packet) Append(b ...byte) *Packet {
	d := p.AppendN(len(b)).Bytes()
	copy(d[len(d)-len(b):], b)
	return p
}

func (p *Packet) AppendN(n int) *Packet {
	if debug.Debug() && n == math.MaxInt {
		slog.Warn("overflow warning", errorx.Trace(nil))
	}

	size := max(n, 0) + len(p.b)
	if cap(p.b) >= size {
		p.b = p.b[:size]
	} else {
		if debug.Debug() {
			slog.Warn("packet memory alloc", errorx.Trace(nil))
		}

		tmp := make([]byte, size, size+DefaulfTail)
		copy(tmp, p.b)
		p.b = tmp
	}
	return p
}

func (p *Packet) Reduce(n int) []byte {
	b := p.Bytes()
	return b[p.ReduceN(n).Data():]
}

func (p *Packet) ReduceTo(to []byte) []byte {
	n := copy(to, p.Reduce(len(to)))
	return to[:n]
}

func (p *Packet) ReduceN(n int) *Packet {
	n = len(p.b) - max(0, n)
	p.b = p.b[:max(n, p.i)]
	return p
}

func (p *Packet) Clone() *Packet {
	n := cap(p.b)
	var c = &Packet{
		b: make([]byte, n),
		i: p.i,
	}
	copy(c.b[:n], p.b[:n])

	return c.Sets(p.Head(), p.Data())
}
