package network

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	netcall "github.com/lysShub/netkit/syscall"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
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

func Test_dosPathCache(t *testing.T) {
	c, err := newDosPathCache()
	require.NoError(t, err)

	fd, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(os.Getpid()))
	require.NoError(t, err)

	var b0 = make([]uint16, 1536)
	n, err := netcall.GetProcessImageFileNameW(fd, b0)
	require.NoError(t, err)
	p0, err := c.Path(b0[:n])
	require.NoError(t, err)

	var b1 = make([]uint16, 1536)
	var size = uint32(len(b1))
	err = netcall.QueryFullProcessImageNameW(fd, 0, b1, &size)
	require.NoError(t, err)
	p1 := windows.UTF16ToString(b1[:size])

	require.Equal(t, p1, p0)
}
