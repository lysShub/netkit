//go:build windows
// +build windows

package network

import (
	"sync"
	"syscall"

	netcall "github.com/lysShub/netkit/syscall"
	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

type network struct {
	tcpMu sync.RWMutex
	tcp   []Elem

	udpMu sync.RWMutex
	udp   []Elem

	*updater
}

func New() (n *network, err error) {
	once.Do(func() {
		global, err = newNetwork()
	})
	if err != nil {
		once = sync.Once{}
		return nil, err
	}
	return global, nil
}

func newNetwork() (*network, error) {
	var n = &network{
		updater: newUpgrader(),
	}
	return n, nil
}

func (n *network) Upgrade(proto uint8) error {
	switch proto {
	case syscall.IPPROTO_TCP, syscall.IPPROTO_UDP, 0:
	default:
		return errors.Errorf("not support protocol %d", proto)
	}
	return n.updater.Upgrade(proto, n)
}

func (n *network) Query(proto uint8, fn func(Elem) (stop bool)) error {
	switch proto {
	case syscall.IPPROTO_TCP, syscall.IPPROTO_UDP, 0:
	default:
		return errors.Errorf("not support protocol %d", proto)
	}

	if proto == 0 || proto == syscall.IPPROTO_TCP {
		n.tcpMu.RLock()
		defer n.tcpMu.RUnlock()
		for i := range n.tcp {
			if fn(n.tcp[i]) {
				return nil
			}
		}
	}

	if proto == 0 || proto == syscall.IPPROTO_UDP {
		n.udpMu.RLock()
		defer n.udpMu.RUnlock()
		for i := range n.udp {
			if fn(n.udp[i]) {
				return nil
			}
		}
	}
	return nil
}

func (n *network) visit(proto uint8, fn func(es []Elem)) error {
	switch proto {
	case syscall.IPPROTO_TCP:
		n.tcpMu.RLock()
		defer n.tcpMu.RUnlock()
		fn(n.tcp)
	case syscall.IPPROTO_UDP:
		n.udpMu.RLock()
		defer n.udpMu.RUnlock()
		fn(n.udp)
	default:
		return errors.Errorf("not support protocol %d", proto)
	}
	return nil
}

type updater struct {
	mu         sync.Mutex
	upgradeing bool

	tcptable netcall.MibTcpTableOwnerPid
	udptable netcall.MibUdpTableOwnerPid

	procPathBuff []uint16

	// todo: maybe conflict
	pidpathCache map[uint32]string

	dosCache *dosPathCache
}

func newUpgrader() *updater {
	var u = &updater{
		procPathBuff: make([]uint16, syscall.MAX_LONG_PATH),
		pidpathCache: map[uint32]string{},
	}

	var err error
	u.dosCache, err = newDosPathCache()
	if err != nil {
		panic(err)
	}
	return u
}

func (u *updater) Upgrade(proto uint8, n *network) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.upgradeing {
		return nil
	} else {
		u.upgradeing = true
		defer func() { u.upgradeing = false }()
	}

	clear(u.pidpathCache)
	switch proto {
	case syscall.IPPROTO_TCP:
		return u.upgradeTCP(n)
	case syscall.IPPROTO_UDP:
		return u.upgradeUDP(n)
	default:
		if err := u.upgradeTCP(n); err != nil {
			return err
		}
		if err := u.upgradeUDP(n); err != nil {
			return err
		}
	}
	return nil
}

func (u *updater) upgradeTCP(n *network) error {
	size := uint32(len(u.tcptable))
	err := netcall.GetExtendedTcpTable(u.tcptable, &size, true)
	if err != nil {
		if errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) {
			u.tcptable = make(netcall.MibTcpTableOwnerPid, size+512)
			return u.upgradeTCP(n)
		}
		return err
	}

	n.tcpMu.Lock()
	defer n.tcpMu.Unlock()
	rows := u.tcptable.Rows()
	for len(n.tcp) < len(rows) {
		n.tcp = append(n.tcp, Elem{})
	}
	n.tcp = n.tcp[:len(rows)]
	for i, e := range rows {
		n.tcp[i].Laddr = e.LocalAddr()
		n.tcp[i].Raddr = e.RemoteAddr()
		n.tcp[i].Proto = syscall.IPPROTO_TCP
		n.tcp[i].Pid = e.OwningPid
		n.tcp[i].Path, err = u.processPath(e.OwningPid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *updater) upgradeUDP(n *network) error {
	size := uint32(len(u.udptable))
	err := netcall.GetExtendedUdpTable(u.udptable, &size, true)
	if err != nil {
		if errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) {
			u.udptable = make(netcall.MibUdpTableOwnerPid, size+512)
			return u.upgradeUDP(n)
		}
		return err
	}

	n.udpMu.Lock()
	defer n.udpMu.Unlock()
	rows := u.udptable.Rows()
	for len(n.udp) < len(rows) {
		n.udp = append(n.udp, Elem{})
	}
	n.udp = n.udp[:len(rows)]
	for i, e := range rows {
		n.udp[i].Laddr = e.LocalAddr()
		n.udp[i].Proto = syscall.IPPROTO_UDP
		n.udp[i].Pid = e.OwningPid
		n.udp[i].Path, err = u.processPath(e.OwningPid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *updater) processPath(pid uint32) (path string, err error) {
	path, has := u.pidpathCache[pid]
	if has {
		return path, nil
	}

	switch pid {
	case 0:
		return SystemIdleName, nil
	case 4:
		return SystemName, nil
	default:
		fd, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if err != nil {
			if err == windows.ERROR_ACCESS_DENIED {
				return SystemProcessName, nil
			} else if err == windows.ERROR_INVALID_PARAMETER {
				// todo: add log
				fd, err = windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, pid)
				if err != nil {
					if err == windows.ERROR_ACCESS_DENIED {
						return SystemProcessName, nil
					}
					return "", errors.Errorf("OpenProcess(%d): %s", pid, err.Error())
				}
			} else {
				return "", errors.Errorf("OpenProcess(%d): %s", pid, err.Error())
			}
		}
		defer windows.CloseHandle(fd)

		if n, err := netcall.GetProcessImageFileNameW(fd, u.procPathBuff); err != nil {
			return "", err
		} else {
			path, err = u.dosCache.Path(u.procPathBuff[:n])
			if err != nil {
				return "", err
			}

			u.pidpathCache[pid] = path
			return path, nil
		}
	}
}

type dosPathCache struct {
	// cache device-dos-name <=> device, eg:"\Device\HarddiskVolume1" <=> 'C'
	cache map[string]byte
}

func newDosPathCache() (*dosPathCache, error) {
	var c = &dosPathCache{
		cache: map[string]byte{},
	}

	var b = make([]uint16, 512)
	for d := 'A'; d < 'Z'; d++ {
		err := netcall.QueryDosDeviceW(string([]rune{d, ':'}), b)
		if err != nil {
			if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
				continue
			}
			return nil, err
		}
		c.cache[windows.UTF16ToString(b)] = byte(d)
	}
	return c, nil
}

func (u *dosPathCache) Path(dosPath []uint16) (string, error) {
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
		return "", errors.Errorf("dos path %s", windows.UTF16ToString(dosPath))
	}

	var dosprefix = windows.UTF16ToString(dosPath[:i])
	d, has := u.cache[dosprefix]
	if !has {
		// todo: maybe update cache, disk hot swapping
		return "", errors.Errorf("not found dos device：%s", windows.UTF16ToString(dosPath))
	}

	dosPath[i-1] = ':'
	dosPath[i-2] = uint16(d)
	return windows.UTF16ToString(dosPath[i-2:]), nil
}
