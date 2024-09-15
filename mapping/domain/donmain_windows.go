//go:build windows
// +build windows

package domain

import (
	"github.com/lysShub/divert-go"
	"github.com/lysShub/netkit/errorx"
	"github.com/lysShub/netkit/packet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

func newCapture() (capture, error) {
	return newDivertCapture()
}

type divertCapture struct {
	handle *divert.Handle

	closeErr errorx.CloseErr
}

func newDivertCapture() (capture, error) {
	if err := divert.Load(divert.DLL); err != nil && !errorx.Temporary(err) {
		return nil, err
	}

	var g = &divertCapture{}
	var err error

	var filter = "inbound and ip and udp and remotePort=53"
	g.handle, err = divert.Open(filter, divert.Network, 0, divert.Sniff|divert.ReadOnly)
	if err != nil {
		return nil, g.close(err)
	}
	return g, nil
}

func (c *divertCapture) close(cause error) error {
	return c.closeErr.Close(func() (errs []error) {
		errs = append(errs, cause)
		if c.handle != nil {
			errs = append(errs, c.handle.Close())
		}
		return errs
	})
}

func (c *divertCapture) Capture(b *packet.Packet) error {
	n, err := c.handle.Recv(b.Bytes(), nil)
	if err != nil {
		return err
	} else if n == 0 {
		return c.Capture(b)
	}
	b.SetData(n)

	ip := header.IPv4(b.Bytes())
	b.DetachN(int(ip.HeaderLength()) + header.UDPMinimumSize)
	return nil
}

func (c *divertCapture) Close() error { return c.close(nil) }
