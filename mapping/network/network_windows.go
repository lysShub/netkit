//go:build windows
// +build windows

package network

import (
	"strconv"
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

	*upgrader
}

func New() *network {
	once.Do(func() {
		global = &network{
			upgrader: newUpgrader(),
		}
	})
	return global
}

func (n *network) Upgrade(proto uint8) error {
	switch proto {
	case syscall.IPPROTO_TCP, syscall.IPPROTO_UDP, 0:
	default:
		return errors.Errorf("not support protocol %d", proto)
	}
	return n.upgrader.Upgrade(proto, n)
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

type upgrader struct {
	mu           sync.Mutex
	tcptable     netcall.MibTcpTableOwnerPid
	udptable     netcall.MibUdpTableOwnerPid
	procpathBuff [syscall.MAX_LONG_PATH]uint16
	pidpathCache map[uint32]string
}

func newUpgrader() *upgrader {
	return &upgrader{
		pidpathCache: map[uint32]string{},
	}
}

func (u *upgrader) Upgrade(proto uint8, n *network) error {
	u.mu.Lock()
	defer u.mu.Unlock()

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

func (u *upgrader) upgradeTCP(n *network) error {
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
		n.tcp[i].Path, err = u.modelPath(e.OwningPid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *upgrader) upgradeUDP(n *network) error {
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
		n.udp[i].Path, err = u.modelPath(e.OwningPid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *upgrader) modelPath(pid uint32) (path string, err error) {
	path, has := n.pidpathCache[pid]
	if has {
		return path, nil
	}

	switch pid {
	case 0:
		path = SystemIdleName
	case 4:
		path = SystemName
	default:
		proc, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if err != nil {
			//  windows.ERROR_ACCESS_DENIED
			return "", errors.WithStack(errors.WithMessage(err, strconv.Itoa(int(pid))))
		}
		defer windows.Close(proc)

		err = windows.GetModuleFileNameEx(proc, 0, &n.procpathBuff[0], uint32(len(n.procpathBuff)))
		if err != nil {
			return "", errors.WithStack(err)
		}
		path = windows.UTF16ToString(n.procpathBuff[:])
	}

	n.pidpathCache[pid] = path
	return path, nil
}
