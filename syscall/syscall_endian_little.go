// copy from encoding\binary\native_endian_little.go
//go:build 386 || amd64 || amd64p32 || alpha || arm || arm64 || loong64 || mipsle || mips64le || mips64p32le || nios2 || ppc64le || riscv || riscv64 || sh || wasm

package syscall

// Hton transport host-byte-order to network-byte-order(big endian)
func Hton[T ~uint16 | ~uint32 | ~uint64](v T) T {
	return ReserveByte(v)
}

// Ntoh transport network-byte-order(big endian) to host-byte-order
func Ntoh[T ~uint16 | ~uint32 | ~uint64](v T) T {
	return ReserveByte(v)
}
