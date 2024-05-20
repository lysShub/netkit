/*
	eth conn
*/

package eth

import "encoding/binary"

func Htons[T ~uint16 | ~uint32 | ~int](p T) T {
	return T(binary.BigEndian.Uint16(
		binary.NativeEndian.AppendUint16(nil, uint16(p)),
	))
}
