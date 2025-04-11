package pid2path

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/lysShub/netkit/debug"
	netcall "github.com/lysShub/netkit/syscall"
	"github.com/lysShub/netkit/test"
	"github.com/stretchr/testify/require"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

type Pid2Path struct {
	// cache device-dos-name <=> device, eg:"\Device\HarddiskVolume1" <=> 'C'
	cache map[string]byte

	systemPath string
}

func New() (*Pid2Path, error) {
	var p2p = &Pid2Path{
		cache:      map[string]byte{},
		systemPath: filepath.Join(os.Getenv("SystemRoot"), ""),
	}

	var b = make([]uint16, 512)
	for d := 'A'; d < 'Z'; d++ {
		err := netcall.QueryDosDeviceW(string([]rune{d, ':'}), b)
		if err != nil {
			if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
				continue
			}
			return nil, errors.WithStack(err)
		}
		p2p.cache[windows.UTF16ToString(b)] = byte(d)
	}

	root := os.Getenv("SYSTEMROOT")
	if root == "" {
		root = `C:\Windows`
	}
	p2p.systemPath = filepath.Join(root, "System32", "ntoskrnl.exe")

	return p2p, nil
}

func (p *Pid2Path) SystemPath() string { return p.systemPath }

func (p *Pid2Path) Path(pid uint32) (path string, err error) {
	path, err = p.path(pid)
	if debug.Debug() && err == nil {
		_, err := os.Stat(path)
		require.NoError(test.T(), err)
	}
	return path, err
}

func (p *Pid2Path) path(pid uint32) (path string, err error) {
	switch pid {
	case 0:
		return "", ErrNotfound(pid) // system-idle
	case 4:
		return p.systemPath, nil
	default:
		fd, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if err != nil {
			if err == windows.ERROR_ACCESS_DENIED {
				return "", ErrNotfound(pid)
			} else {
				return "", errors.WithStack(errors.WithMessage(err, strconv.Itoa(int(pid))))
			}
		}
		defer windows.CloseHandle(fd)

		namePtr := getName()
		name := *namePtr
		defer func() {
			*namePtr = name
			putName(namePtr)
		}()

	start:
		// GetProcessImageFileNameW 相比 QueryFullProcessImageNameW 有更高性能, 且兼容性更好
		if n, err := netcall.GetProcessImageFileNameW(fd, name); err != nil {
			if errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) {
				name = append(name, 0)
				name = name[:cap(name)]
				goto start
			} else {
				return "", err
			}
		} else if n == 0 {
			return p.systemPath, nil // eg: Secure System
		} else {
			name = name[:n]

			const (
				Registry = "Registry"
				vmmem    = "vmmem"
			)
			switch len(name) {
			case len(Registry):
				if windows.UTF16ToString(name) == Registry {
					return p.systemPath, nil
				}
			case len(vmmem):
				if windows.UTF16ToString(name) == vmmem {
					return "", ErrNotfound(pid)
				}
			default:
			}
			return p.dosPath(name)
		}
	}
}

func (p *Pid2Path) dosPath(dosPath []uint16) (path string, err error) {
	// get device dos path, eg: \Device\HarddiskVolume1
	var n, i int
	const slash = '\\'
	for ; i < len(dosPath); i++ {
		if dosPath[i] == slash {
			if n += 1; n == 3 {
				break
			}
		}
	}
	if i < 5 {
		return "", errors.Errorf("invalid dos path %s", windows.UTF16ToString(dosPath))
	}

	var dosprefix = windows.UTF16ToString(dosPath[:i])
	d, has := p.cache[dosprefix]
	if !has {
		return "", errors.Errorf("not found dos device %s", windows.UTF16ToString(dosPath))
	}

	dosPath[i-1] = ':'
	dosPath[i-2] = uint16(d)
	return windows.UTF16ToString(dosPath[i-2:]), nil
}

var (
	nameBufSize atomic.Uint32
	nameBufPool = sync.Pool{
		New: func() any {
			var b = make([]uint16, nameBufSize.Load())
			return &b
		},
	}
)

func getName() *[]uint16 {
	return nameBufPool.Get().(*[]uint16)
}
func putName(name *[]uint16) {
	if name == nil {
		return
	} else {
		n := uint32(cap(*name))
		*name = (*name)[:n]
		if nameBufSize.Load() < n {
			nameBufSize.Store(n)
		}
		nameBufPool.Put(name)
	}
}

func SeDebugPrivilege() error {
	var token windows.Token
	if err := windows.OpenProcessToken(
		windows.CurrentProcess(),
		windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY,
		&token,
	); err != nil {
		return errors.WithStack(err)
	}
	defer token.Close()

	var luid windows.LUID
	err := windows.LookupPrivilegeValue(nil, windows.StringToUTF16Ptr("SeDebugPrivilege"), &luid)
	if err != nil {
		return errors.WithStack(err)
	}

	privileges := &windows.Tokenprivileges{
		PrivilegeCount: 1,
		Privileges: [1]windows.LUIDAndAttributes{{
			Luid:       luid,
			Attributes: windows.SE_PRIVILEGE_ENABLED,
		}},
	}
	err = windows.AdjustTokenPrivileges(token, false, privileges, 0, nil, nil)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
