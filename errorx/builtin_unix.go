//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris zos

package errorx

import (
	"errors"

	"golang.org/x/sys/unix"
)

var ErrConnectRefused = unix.ECONNREFUSED

func connectRefused(err error) bool {
	return errors.Is(err, unix.ECONNREFUSED)
}

var ErrNetworkUnreach = unix.ENETUNREACH

func networkUnreach(err error) bool {
	return errors.Is(err, unix.ENETUNREACH)
}

var ErrBuffTooSmall = unix.ENOBUFS

func buffTooSmall(err error) bool {
	return errors.Is(err, unix.ENOBUFS)
}

var ErrAddrNotAvail = unix.EADDRNOTAVAIL

func addrNotAvail(err error) bool {
	return errors.Is(err, unix.EADDRNOTAVAIL)
}
