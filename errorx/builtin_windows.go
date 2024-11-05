//go:build windows
// +build windows

package errorx

import (
	"errors"

	"golang.org/x/sys/windows"
)

func connectRefused(err error) bool {
	return errors.Is(err, windows.WSAECONNREFUSED) ||
		errors.Is(err, windows.ERROR_CONNECTION_REFUSED)
}

func netUnreach(err error) bool {
	return errors.Is(err, windows.WSAENETUNREACH) ||
		errors.Is(err, windows.ERROR_NETWORK_UNREACHABLE)
}
