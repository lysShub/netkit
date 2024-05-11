//go:build linux
// +build linux

package process

import (
	"fmt"

	"golang.org/x/sys/unix"
)

type mapping struct{}

func newMapping() (Mapping, error) {
	panic("not implement")
}

func protoStr(proto uint8) string {
	switch proto {
	case unix.IPPROTO_TCP:
		return "tcp"
	case unix.IPPROTO_UDP:
		return "udp"
	case unix.IPPROTO_ICMP:
		return "icmp"
	case unix.IPPROTO_ICMPV6:
		return "icmp6"
	default:
		return fmt.Sprintf("unknown(%d)", int(proto))
	}
}
