package pid2path

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/lysShub/netkit/errorx"
	netcall "github.com/lysShub/netkit/syscall"
	"github.com/mitchellh/go-ps"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

/*
	goos: windows
	goarch: amd64
	pkg: github.com/lysShub/netkit/pid2path
	cpu: Intel(R) Xeon(R) CPU E5-1650 v4 @ 3.60GHz
	Benchmark_GetProcessImageFileNameW-12            1778310             669.5 ns/op
	Benchmark_QueryFullProcessImageNameW-12            44554             25820 ns/op
	Benchmark_GetModuleFileNameEx-12                   46960             25487 ns/op
	PASS
	ok      github.com/lysShub/netkit/pid2path      5.763s
*/

func Benchmark_GetProcessImageFileNameW(b *testing.B) {
	var pathBuff = make([]uint16, syscall.MAX_LONG_PATH)
	fd, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(os.Getpid()))
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		size, err := netcall.GetProcessImageFileNameW(fd, pathBuff)
		if err != nil {
			require.NoError(b, err)
		}
		_ = windows.UTF16ToString(pathBuff[:size])
	}
}

func Benchmark_QueryFullProcessImageNameW(b *testing.B) {
	var pathBuff = make([]uint16, syscall.MAX_LONG_PATH)
	fd, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(os.Getpid()))
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		size := uint32(len(pathBuff))
		err := netcall.QueryFullProcessImageNameW(fd, 0, pathBuff, &size)
		if err != nil {
			require.NoError(b, err)
		}
		_ = windows.UTF16ToString(pathBuff[:size])
	}
}

func Benchmark_GetModuleFileNameEx(b *testing.B) {
	var pathBuff = make([]uint16, syscall.MAX_LONG_PATH)
	fd, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(os.Getpid()))
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		err := windows.GetModuleFileNameEx(fd, 0, &pathBuff[0], uint32(len(pathBuff)))
		if err != nil {
			require.NoError(b, err)
		}
		_ = windows.UTF16ToString(pathBuff)
	}
}

func Test_Pid2Path(t *testing.T) {
	t.Run("DebugPrivilege", func(t *testing.T) {
		list, err := ps.Processes()
		require.NoError(t, err)

		require.NoError(t, SeDebugPrivilege())
		p2p, err := New()
		require.NoError(t, err)
		for _, p := range list {
			pid := uint32(p.Pid())
			path, err := p2p.Path(uint32(pid))
			if err != nil {
				require.True(t, errorx.NotFound(err))
				require.Contains(t, []string{"vmmem", "[System Process]"}, p.Executable())
				continue
			}

			exp := p.Executable()
			act := filepath.Base(path)
			switch exp {
			case "System", "Secure System", "Registry":
				require.Equal(t, p2p.SystemPath(), path)
			default:
				require.Equal(t, exp, act)
			}
		}

	})

	t.Run("normal", func(t *testing.T) {
		list, err := ps.Processes()
		require.NoError(t, err)

		require.NoError(t, SeDebugPrivilege())
		p2p, err := New()
		require.NoError(t, err)
		for _, p := range list {
			pid := uint32(p.Pid())
			path, err := p2p.Path(uint32(pid))
			if err != nil {
				require.True(t, errorx.NotFound(err))
				continue
			}

			exp := p.Executable()
			act := filepath.Base(path)
			switch exp {
			case "System", "Secure System", "Registry":
				require.Equal(t, p2p.SystemPath(), path)
			default:
				require.Equal(t, exp, act)
			}
		}
	})
}
