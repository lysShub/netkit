//go:build !windows && !linux
// +build !windows,!linux

package route

func GetTable() (Table, error) {
	panic("not support")
}
