package syscall

import (
	"slices"
	"unsafe"
)

func ReserveByte[T ~uint16 | ~uint32 | ~uint64](v T) T {
	n := unsafe.Sizeof(v)

	s := unsafe.Slice(
		(*byte)(unsafe.Pointer(&v)), n,
	)
	slices.Reverse(s)

	return *(*T)(unsafe.Pointer(unsafe.SliceData(s)))
}
