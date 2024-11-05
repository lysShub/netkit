//go:build windows
// +build windows

package errorx

import (
	"errors"

	"golang.org/x/sys/windows"
)

func connectRefused(err error) bool {
	return errors.Is(err, windows.WSAECONNREFUSED)
}

func netUnreach(err error) bool {
	return errors.Is(err, windows.WSAENETUNREACH)
}
