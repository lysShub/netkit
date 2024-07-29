//go:build linux
// +build linux

package route

import (
	"syscall"

	netcall "github.com/lysShub/netkit/syscall"
)

type EntryRaw []syscall.NetlinkRouteAttr

func (e *Entry) Name() (string, error) {
	return netcall.IoctlGifname(int(e.Interface))
}
