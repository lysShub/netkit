//go:build !windows
// +build !windows

package domain

import "github.com/pkg/errors"

func newCapture() (Capture, error) {
	return nil, errors.New("not support")
}

func cleanupDnsCache() {}
