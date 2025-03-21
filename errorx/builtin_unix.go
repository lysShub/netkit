//go:build unix
// +build unix

package errorx

import (
	"errors"

	"golang.org/x/sys/unix"
)

var ErrConnectReset = unix.ECONNRESET

func connectResed(err error) bool {
	return errors.Is(err, unix.ECONNRESET)
}

var ErrConnectRefused = unix.ECONNREFUSED

func connectRefused(err error) bool {
	return errors.Is(err, unix.ECONNREFUSED)
}

var ErrConnectAborted = unix.ECONNABORTED

func connectAborted(err error) bool {
	return errors.Is(err, unix.ECONNABORTED)
}

var ErrNetworkUnreach = unix.ENETUNREACH

func networkUnreach(err error) bool {
	return errors.Is(err, unix.ENETUNREACH) || errors.Is(err, unix.EHOSTUNREACH)
}

var ErrBuffTooSmall = unix.ENOBUFS

func buffTooSmall(err error) bool {
	return errors.Is(err, unix.ENOBUFS)
}

var ErrAddrNotAvail = unix.EADDRNOTAVAIL

func addrNotAvail(err error) bool {
	return errors.Is(err, unix.EADDRNOTAVAIL)
}

var ErrNetTimeout = unix.ETIMEDOUT

func netTimeout(err error) bool {
	return errors.Is(err, ErrNetTimeout)
}
