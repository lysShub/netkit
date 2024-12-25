//go:build windows
// +build windows

package errorx

import (
	"errors"

	"golang.org/x/sys/windows"
)

var ErrConnectReset = windows.WSAECONNRESET

func connectResed(err error) bool {
	return errors.Is(err, windows.WSAECONNRESET)
}

var ErrConnectRefused = windows.WSAECONNREFUSED

func connectRefused(err error) bool {
	return errors.Is(err, windows.WSAECONNREFUSED) ||
		errors.Is(err, windows.ERROR_CONNECTION_REFUSED)
}

var ErrNetworkUnreach = windows.WSAENETUNREACH

func networkUnreach(err error) bool {
	return errors.Is(err, windows.WSAENETUNREACH) ||
		errors.Is(err, windows.ERROR_NETWORK_UNREACHABLE)
}

var ErrBuffTooSmall = windows.ERROR_INSUFFICIENT_BUFFER

func buffTooSmall(err error) bool {
	return errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) ||
		errors.Is(err, windows.ERROR_DS_USER_BUFFER_TO_SMALL)
}

var ErrAddrNotAvail = windows.WSAEADDRNOTAVAIL

func addrNotAvail(err error) bool {
	return errors.Is(err, windows.WSAEADDRNOTAVAIL)
}
