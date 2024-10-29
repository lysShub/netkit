//go:build windows
// +build windows

package domain

import (
	"os/exec"

	"github.com/lysShub/divert-go"
	"github.com/lysShub/netkit/errorx"
)

func newCapture() (Capture, error) {
	return newDivertCapture()
}

type divertCapture struct {
	handle *divert.Handle

	closeErr errorx.CloseErr
}

func newDivertCapture() (Capture, error) {
	if err := divert.Load(divert.DLL); err != nil && !errorx.Temporary(err) {
		return nil, err
	}

	var g = &divertCapture{}
	var err error

	var filter = "inbound and ip and remotePort=53"
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

func (c *divertCapture) Capture(ip []byte) (int, error) {
	n, err := c.handle.Recv(ip, nil)
	if err != nil {
		return 0, err
	} else {
		return n, err
	}
}

func (c *divertCapture) Close() error { return c.close(nil) }

func cleanupDnsCache() {
	exec.Command("ipconfig", "/flushdns").CombinedOutput()
}
