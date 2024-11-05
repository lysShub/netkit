//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris zos

package errorx

import (
	"errors"

	"golang.org/x/sys/unix"
)

func connectRefused(err error) bool {
	return errors.Is(err, unix.ECONNREFUSED)
}

func netUnreach(err error) bool {
	return errors.Is(err, unix.ENETUNREACH)
}
