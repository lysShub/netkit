/*
	eth conn
*/

package eth

import (
	"encoding/binary"
	"net"

	"github.com/lysShub/netkit/route"
	"github.com/mdlayher/arp"
	"github.com/pkg/errors"
)

func Htons[T ~uint16 | ~uint32 | ~int](p T) T {
	return T(binary.BigEndian.Uint16(
		binary.NativeEndian.AppendUint16(make([]byte, 0, 2), uint16(p)),
	))
}

func Gateway() (net.HardwareAddr, error) {
	t, err := route.GetTable()
	if err != nil {
		return nil, err
	} else if !t[0].Next.IsValid() {
		return nil, errors.New("can't get default route next hop")
	}

	ifi, err := net.InterfaceByIndex(int(t[0].Interface))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	c, err := arp.Dial(ifi)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	hw, err := c.Resolve(t[0].Next)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hw, nil
}
