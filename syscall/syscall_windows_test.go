package syscall_test

import (
	"errors"
	"fmt"
	"testing"

	netcall "github.com/lysShub/netkit/syscall"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
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

func Test_QueryDosDeviceW(t *testing.T) {

	var b = make([]uint16, 0xffff)

	err := netcall.QueryDosDeviceW("A:", b)
	require.NoError(t, err)

	// windows.ERROR_FLT_FILTER_NOT_FOUND

	str := windows.UTF16ToString(b)
	fmt.Println(str)

	// var ss []byte
	// for _, e := range b {
	// 	if e > 0xff {
	// 		panic("")
	// 	}
	// 	ss = append(ss, byte(e))
	// }

	// str := string(ss)
	// fmt.Println(str)

}

func Test_IfIndexToName(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		n, err := netcall.IfIndexToName(2)
		require.NoError(t, err)
		require.Greater(t, len(n), 0, n)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := netcall.IfIndexToName(111)
		require.True(t, errors.Is(err, windows.ERROR_FILE_NOT_FOUND))
	})

}

func Test_ConvertInterfaceLuidToAlias(t *testing.T) {

	t.Run("normal", func(t *testing.T) {
		id, err := netcall.ConvertInterfaceIndexToLuid(2)
		require.NoError(t, err)

		n, err := netcall.ConvertInterfaceLuidToAlias(uint64(id))
		require.NoError(t, err)
		require.Greater(t, len(n), 0, n)
	})

	t.Run("not found", func(t *testing.T) {
		var id uint64 = 12345678

		_, err := netcall.ConvertInterfaceLuidToAlias(id)
		require.True(t, errors.Is(err, windows.ERROR_FILE_NOT_FOUND))
	})
}
