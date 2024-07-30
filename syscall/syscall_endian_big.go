// copy from encoding\binary\native_endian_big.go
//go:build armbe || arm64be || m68k || mips || mips64 || mips64p32 || ppc || ppc64 || s390 || s390x || shbe || sparc || sparc64

package syscall

import "golang.org/x/exp/constraints"

// Hton transport host-byte-order to network-byte-order(big endian)
func Hton[T constraints.Float | constraints.Integer](v T) T {
	return v
}

// Ntoh transport network-byte-order(big endian) to host-byte-order
func Ntoh[T constraints.Float | constraints.Integer](v T) T {
	return v
}
