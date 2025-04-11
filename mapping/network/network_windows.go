//go:build windows
// +build windows

package network

import (
	"sync"
	"syscall"

	"github.com/lysShub/netkit/pid2path"
	netcall "github.com/lysShub/netkit/syscall"
	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

type Network struct {
	tcpMu sync.RWMutex
	tcp   []Elem

	udpMu sync.RWMutex
	udp   []Elem

	*updater
}

func New() (n *Network, err error) {
	once.Do(func() {
		global, err = newNetwork()
	})
	if err != nil {
		once = sync.Once{}
		return nil, err
	}
	return global, nil
}

func newNetwork() (*Network, error) {
	var n = &Network{
		updater: newUpgrader(),
	}
	return n, nil
}

func (n *Network) Upgrade(proto uint8) error {
	switch proto {
	case syscall.IPPROTO_TCP, syscall.IPPROTO_UDP, 0:
	default:
		return errors.Errorf("not support protocol %d", proto)
	}
	return n.updater.Upgrade(proto, n)
}

func (n *Network) Query(proto uint8, fn func(Elem) (stop bool)) error {
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

func (n *Network) visit(proto uint8, fn func(es []Elem)) error {
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

	// todo: maybe conflict
	pidpathCache map[uint32]string

	pid2path *pid2path.Pid2Path
}

func newUpgrader() *updater {
	var u = &updater{
		pidpathCache: map[uint32]string{},
	}

	var err error
	if u.pid2path, err = pid2path.New(); err != nil {
		panic(err)
	}
	return u
}

func (u *updater) Upgrade(proto uint8, n *Network) error {
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

func (u *updater) upgradeTCP(n *Network) error {
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

func (u *updater) upgradeUDP(n *Network) error {
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

	path, err = u.pid2path.Path(pid)
	if err != nil {
		return "", err
	}
	u.pidpathCache[pid] = path
	return path, nil
}
