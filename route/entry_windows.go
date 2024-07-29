//go:build windows
// +build windows

package route

import (
	netcall "github.com/lysShub/netkit/syscall"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

type EntryRaw = winipcfg.MibIPforwardRow2

func (e *Entry) Name() (string, error) {
	return netcall.ConvertInterfaceLuidToAlias(uint64(e.raw.InterfaceLUID))
}
