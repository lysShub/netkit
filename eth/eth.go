/*
	eth conn
*/

package eth

import "encoding/binary"

func Htons(b uint16) uint16 {
	return binary.BigEndian.Uint16(
		binary.NativeEndian.AppendUint16(nil, b),
	)
}
