/*
	eth conn
*/

package eth

import (
	"encoding/binary"
)

func Htons[T ~uint16 | ~uint32 | ~int](p T) T {
	return T(binary.BigEndian.Uint16(
		binary.NativeEndian.AppendUint16(make([]byte, 0, 2), uint16(p)),
	))
}
