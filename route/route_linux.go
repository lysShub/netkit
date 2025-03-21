//go:build linux
// +build linux

package route

import (
	"net"
	"net/netip"
	"syscall"
	"unsafe"

	netcall "github.com/lysShub/netkit/syscall"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

func (e *Entry) HardwareAddr() (net.HardwareAddr, error) {
	name, err := e.Name()
	if err != nil {
		return nil, err
	}
	return netcall.IoctlGifhwaddr(name)
}

// GetTable get ipv4 route entries
func GetTable() (Table, error) {
	// todo: set socket timeout
	tab, err := syscall.NetlinkRIB(unix.RTM_GETROUTE, unix.AF_INET)
	if err != nil {
		return nil, err
	}
	msgs, err := syscall.ParseNetlinkMessage(tab)
	if err != nil {
		return nil, err
	}

	var table Table
	for i := 0; i < len(msgs); i++ {
		m := msgs[i]
		switch m.Header.Type {
		case unix.RTM_NEWROUTE:
			attrs, err := syscall.ParseNetlinkRouteAttr(&m)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			rt := (*unix.RtMsg)(unsafe.Pointer(unsafe.SliceData(m.Data)))
			e := collectEntry(attrs, rt.Dst_len)
			if !e.Dest.IsValid() {
				e.Dest = netip.PrefixFrom(netip.IPv4Unspecified(), 0)
			}
			if !e.Addr.IsValid() {
				name, err := netcall.IoctlGifname(int(e.Interface))
				if err != nil {
					return nil, err
				}
				addr, err := netcall.IoctlGifaddr(name)
				if err != nil {
					return nil, err
				}
				e.Addr = addr.Addr()
			}
			e.raw = attrs

			table = append(table, e)
		case unix.NLMSG_DONE:
			i = len(msgs) // break
		case unix.NLMSG_NOOP:
			continue
		case unix.NLMSG_ERROR:
			rt := (*unix.NlMsgerr)(unsafe.Pointer(unsafe.SliceData(m.Data)))
			return nil, errors.WithStack(unix.Errno(-rt.Error))
		default:
			return nil, errors.Errorf("unexpect nlmsghdr type 0x%02x", m.Header.Type)
		}
	}

	if len(table) == 0 {
		return nil, errors.New("can't get route table")
	}
	table.Sort()
	return table, nil
}

func collectEntry(attrs []syscall.NetlinkRouteAttr, ones uint8) Entry {
	var e Entry
	for _, attr := range attrs {
		switch attr.Attr.Type {
		case unix.RTA_GATEWAY:
			e.Next, _ = netip.AddrFromSlice(attr.Value)
		case unix.RTA_DST:
			addr, ok := netip.AddrFromSlice(attr.Value)
			if ok {
				e.Dest = netip.PrefixFrom(addr, int(ones))
			}
		case unix.RTA_SRC, unix.RTA_PREFSRC:
			e.Addr, _ = netip.AddrFromSlice(attr.Value)
		case unix.RTA_OIF:
			idx := *(*int32)(unsafe.Pointer(unsafe.SliceData(attr.Value)))
			e.Interface = uint32(idx)
		case unix.RTA_PRIORITY: // unix.RTA_METRICS
			metric := *(*int32)(unsafe.Pointer(unsafe.SliceData(attr.Value)))
			e.Metric = uint32(metric)
		}
	}
	return e
}
