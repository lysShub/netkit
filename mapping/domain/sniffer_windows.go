//go:build windows
// +build windows

package domain

import (
	"os/exec"

	"github.com/lysShub/divert-go"
	"github.com/lysShub/netkit/errorx"
)

func newSniffer() (Sniffer, error) {
	return newDivert()
}

type divertSniffer struct {
	handle *divert.Handle

	closeErr errorx.CloseErr
}

func newDivert() (Sniffer, error) {
	if err := divert.Load(divert.DLL); err != nil && !errorx.IsTemporary(err) {
		return nil, err
	}

	var g = &divertSniffer{}
	var err error

	var filter = "inbound and ip and remotePort=53"
	g.handle, err = divert.Open(filter, divert.Network, 0, divert.Sniff|divert.ReadOnly)
	if err != nil {
		return nil, g.close(err)
	}
	return g, nil
}

func (c *divertSniffer) close(cause error) error {
	return c.closeErr.Close(func() (errs []error) {
		errs = append(errs, cause)
		if c.handle != nil {
			errs = append(errs, c.handle.Close())
		}
		return errs
	})
}

func (c *divertSniffer) Sniff(ip []byte) (int, error) {
	n, err := c.handle.Recv(ip, nil)
	if err != nil {
		return 0, err
	} else {
		return n, err
	}
}

func (c *divertSniffer) Close() error { return c.close(nil) }

func cleanupDnsCache() {
	exec.Command("ipconfig", "/flushdns").CombinedOutput()
}
