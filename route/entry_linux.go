//go:build linux
// +build linux

package route

import netcall "github.com/lysShub/netkit/syscall"

type EntryRaw struct{}

func (e *Entry) Name() (string, error) {
	return netcall.IoctlGifname(int(e.Interface))
}
