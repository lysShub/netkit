package syscall_test

import (
	"fmt"
	"testing"

	netcall "github.com/lysShub/netkit/syscall"
	"github.com/stretchr/testify/require"
)

func Benchmark_GetExtendedTcpTable_Norder(b *testing.B) {
	var size uint32 = 65536
	var table = make(netcall.MibTcpTableOwnerPid, size)

	for i := 0; i < b.N; i++ {
		size = uint32(len(table))
		err := netcall.GetExtendedTcpTable(table, &size, false)
		if err != nil {
			panic(err)
		}
	}
}

func Benchmark_GetExtendedTcpTable_Order(b *testing.B) {
	var size uint32 = 65536
	var table = make(netcall.MibTcpTableOwnerPid, size)

	for i := 0; i < b.N; i++ {
		size = uint32(len(table))
		err := netcall.GetExtendedTcpTable(table, &size, true)
		if err != nil {
			panic(err)
		}
	}
}

func Test_GetExtendedTcpTable(t *testing.T) {

	var size uint32 = 65536
	var table = make(netcall.MibTcpTableOwnerPid, size)

	err := netcall.GetExtendedTcpTable(table, &size, true)
	require.NoError(t, err)

	rows := table.Rows()

	for _, e := range rows {
		fmt.Println(e.LocalAddr().String(), e.OwningPid)
	}
}

func Test_GetExtendedUdpTable(t *testing.T) {
	var size uint32 = 65536
	var table = make(netcall.MibUdpTableOwnerPid, size)

	err := netcall.GetExtendedUdpTable(table, &size, true)
	require.NoError(t, err)

	rows := table.Rows()

	for _, e := range rows {
		fmt.Println(e.LocalAddr().String())
	}
}
