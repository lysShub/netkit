package network

import (
	"fmt"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Network(t *testing.T) {
	n, err := New()
	require.NoError(t, err)

	require.NoError(t, n.Upgrade(0))

	n.Query(syscall.IPPROTO_TCP, func(e Elem) (stop bool) {
		fmt.Println(e.Laddr.String(), filepath.Base(e.Path))
		return false
	})

}

func TestXxx(t *testing.T) {

	u := newUpgrader()

	var d = []uint16{}
	for _, e := range `\Device\HarddiskVolume2\Windows` {
		d = append(d, uint16(e))
	}

	a, err := u.devicePath(d)
	require.NoError(t, err)

	fmt.Println(a)

}
