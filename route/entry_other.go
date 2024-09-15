//go:build !windows && !linux
// +build !windows,!linux

package route

type EntryRaw struct{}

func (e *Entry) Name() (string, error) {
	panic("not support")
}
