//go:build !windows
// +build !windows

package domain

func newCapture() (capture, error) {
	panic("not support")
}
