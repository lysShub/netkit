//go:build windows
// +build windows

package syscall

import (
	"encoding/binary"
	"net"
	"net/netip"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

var (
	iphlpapi                        = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetIpForwardTable           = iphlpapi.NewProc("GetIpForwardTable")
	procGetIpAddrTable              = iphlpapi.NewProc("GetIpAddrTable")
	procGetIpForwardEntry2          = iphlpapi.NewProc("GetIpForwardEntry2")
	procGetIpInterfaceEntry         = iphlpapi.NewProc("GetIpInterfaceEntry")
	procGetExtendedTcpTable         = iphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable         = iphlpapi.NewProc("GetExtendedUdpTable")
	procIfIndexToName               = iphlpapi.NewProc("if_indextoname")
	procConvertInterfaceLuidToAlias = iphlpapi.NewProc("ConvertInterfaceLuidToAlias")

	kernel32                       = windows.NewLazySystemDLL("kernel32.dll")
	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
	procQueryDosDeviceW            = kernel32.NewProc("QueryDosDeviceW")

	psapi                        = windows.NewLazySystemDLL("psapi.dll")
	procGetProcessImageFileNameW = psapi.NewProc("GetProcessImageFileNameW")
)

// GetIpForwardTable get sorted ip route entries
func GetIpForwardTable(table MibIpForwardTable, size *uint32, order bool) error {
	r1, _, _ := syscall.SyscallN(
		procGetIpForwardTable.Addr(),
		uintptr(unsafe.Pointer(unsafe.SliceData(table))),
		uintptr(unsafe.Pointer(size)),
		intbool(order),
	)
	if r1 != 0 {
		return errors.WithStack(syscall.Errno(r1))
	}
	return nil
}

func intbool(v bool) uintptr {
	if v {
		return 1
	}
	return 0
}

type MibIpForwardTable []byte

func (m MibIpForwardTable) NumEntries() uint32 {
	return binary.NativeEndian.Uint32(m[:4])
}

func (m MibIpForwardTable) MibForwardRows() []MibIpForwardRow {
	return unsafe.Slice(
		(*MibIpForwardRow)(unsafe.Pointer(unsafe.SliceData(m[4:]))),
		m.NumEntries(),
	)
}

// MIB_IPFORWARDROW
type MibIpForwardRow struct {
	dest      uint32
	mask      uint32
	Policy    uint32
	nextHop   uint32
	IfIndex   uint32
	Type      uint32
	Proto     uint32
	Age       uint32
	NextHopAS uint32
	Metric1   uint32
	Metric2   uint32
	Metric3   uint32
	Metric4   uint32
	Metric5   uint32
}

func (r MibIpForwardRow) DestAddr() netip.Prefix {
	m := *(*[4]byte)(unsafe.Pointer(&r.dest))
	ones, _ := net.IPMask(m[:]).Size()

	return netip.PrefixFrom(r.destAddr(), ones)
}

func (r MibIpForwardRow) destAddr() netip.Addr {
	return netip.AddrFrom4(*(*[4]byte)(unsafe.Pointer(&r.dest)))
}

func (r MibIpForwardRow) NextHop() netip.Addr {
	a := binary.NativeEndian.AppendUint32(nil, r.nextHop)
	return netip.AddrFrom4([4]byte(a))
}

func GetIpAddrTable(table MibIpAddrTable, size *uint32, order bool) error {
	var pTable, intbool uintptr
	if order {
		intbool = 1
	}
	if len(table) > 0 {
		pTable = uintptr(unsafe.Pointer(unsafe.SliceData(table)))
	}

	r1, _, _ := syscall.SyscallN(
		procGetIpAddrTable.Addr(),
		pTable,
		uintptr(unsafe.Pointer(size)),
		intbool,
	)
	if r1 != 0 {
		return errors.WithStack(syscall.Errno(r1))
	}
	return nil
}

type MibIpAddrTable []byte

func (m MibIpAddrTable) NumEntries() uint32 {
	return binary.NativeEndian.Uint32(m[:4])
}

func (m MibIpAddrTable) Rows() []MibIpAddrRow {
	return unsafe.Slice(
		(*MibIpAddrRow)(unsafe.Pointer(unsafe.SliceData(m[4:]))),
		m.NumEntries(),
	)
}

// https://learn.microsoft.com/en-us/windows/win32/api/ipmib/ns-ipmib-mib_ipaddrrow_w2k
type MibIpAddrRow struct {
	addr      uint32
	Index     uint32
	mask      uint32
	BCastAddr uint32
	ReasmSize uint32
	unused1   uint16
	unused2   uint16
}

func (r MibIpAddrRow) Addr() netip.Prefix {
	addr := netip.AddrFrom4(*(*[4]byte)(unsafe.Pointer(&r.addr)))

	m := *(*[4]byte)(unsafe.Pointer(&r.mask))
	ones, _ := net.IPMask(m[:]).Size()

	return netip.PrefixFrom(addr, ones)
}

func GetIpForwardEntry2(row *winipcfg.MibIPforwardRow2) error {
	r1, _, _ := syscall.SyscallN(
		procGetIpForwardEntry2.Addr(),
		uintptr(unsafe.Pointer(row)),
	)
	if r1 == windows.NO_ERROR {
		return nil
	}
	return errors.WithStack(syscall.Errno(r1))
}

func GetIpInterfaceEntry(entry *winipcfg.MibIPInterfaceRow) error {
	r1, _, _ := syscall.SyscallN(
		procGetIpInterfaceEntry.Addr(),
		uintptr(unsafe.Pointer(entry)),
	)
	if r1 == windows.NO_ERROR {
		return nil
	}
	return errors.WithStack(syscall.Errno(r1))
}

// https://learn.microsoft.com/en-us/windows/win32/api/tcpmib/ns-tcpmib-mib_tcptable_owner_pid
type MibTcpTableOwnerPid []byte

func (m MibTcpTableOwnerPid) NumEntries() uint32 {
	return binary.NativeEndian.Uint32(m[:4])
}

func (m MibTcpTableOwnerPid) Rows() []MibTcpRowOwnerPid {
	return unsafe.Slice(
		(*MibTcpRowOwnerPid)(unsafe.Pointer(unsafe.SliceData(m[4:]))),
		m.NumEntries(),
	)
}

// https://learn.microsoft.com/en-us/windows/win32/api/tcpmib/ns-tcpmib-mib_tcprow_owner_pid
type MibTcpRowOwnerPid struct {
	State      uint32
	localAddr  uint32
	localPort  uint32
	remoteAddr uint32
	remotePort uint32
	OwningPid  uint32
}

func (m MibTcpRowOwnerPid) LocalAddr() netip.AddrPort {
	return netip.AddrPortFrom(
		netip.AddrFrom4(*(*[4]byte)(unsafe.Pointer(&m.localAddr))),
		Ntoh(uint16(m.localPort)), // ntohs
	)
}

func (m MibTcpRowOwnerPid) RemoteAddr() netip.AddrPort {
	if m.remotePort == 0 {
		return netip.AddrPort{}
	}
	return netip.AddrPortFrom(
		netip.AddrFrom4(*(*[4]byte)(unsafe.Pointer(&m.remoteAddr))),
		Ntoh(uint16(m.remotePort)), // ntohs
	)
}

// GetExtendedTcpTable only support ipv4 TCP_TABLE_OWNER_PID_ALL
// https://learn.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getextendedtcptable
func GetExtendedTcpTable(table MibTcpTableOwnerPid, size *uint32, order bool) error {
	const (
		TCP_TABLE_BASIC_LISTENER = iota
		TCP_TABLE_BASIC_CONNECTIONS
		TCP_TABLE_BASIC_ALL
		TCP_TABLE_OWNER_PID_LISTENER
		TCP_TABLE_OWNER_PID_CONNECTIONS
		TCP_TABLE_OWNER_PID_ALL
		TCP_TABLE_OWNER_MODULE_LISTENER
		TCP_TABLE_OWNER_MODULE_CONNECTIONS
		TCP_TABLE_OWNER_MODULE_ALL
	)

	r1, _, _ := syscall.SyscallN(
		procGetExtendedTcpTable.Addr(),
		uintptr(unsafe.Pointer(unsafe.SliceData(table))),
		uintptr(unsafe.Pointer(size)),
		intbool(order),
		uintptr(syscall.AF_INET),
		TCP_TABLE_OWNER_PID_ALL,
		0,
	)
	if r1 == windows.NO_ERROR {
		return nil
	}
	return errors.WithStack(syscall.Errno(r1))
}

type MibUdpTableOwnerPid []byte

func (m MibUdpTableOwnerPid) NumEntries() uint32 {
	return binary.NativeEndian.Uint32(m[:4])
}

func (m MibUdpTableOwnerPid) Rows() []MibUdpRowOwnerPid {
	return unsafe.Slice(
		(*MibUdpRowOwnerPid)(unsafe.Pointer(unsafe.SliceData(m[4:]))),
		m.NumEntries(),
	)
}

type MibUdpRowOwnerPid struct {
	localAddr uint32
	localPort uint32
	OwningPid uint32
}

func (m MibUdpRowOwnerPid) LocalAddr() netip.AddrPort {
	return netip.AddrPortFrom(
		netip.AddrFrom4(*(*[4]byte)(unsafe.Pointer(&m.localAddr))),
		Ntoh(uint16(m.localPort)), // ntohs
	)
}

// GetExtendedUdpTable only support ipv4 UDP_TABLE_OWNER_PID
// https://learn.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getextendedudptable
func GetExtendedUdpTable(table MibUdpTableOwnerPid, size *uint32, order bool) error {
	const (
		UDP_TABLE_BASIC uintptr = iota
		UDP_TABLE_OWNER_PID
		UDP_TABLE_OWNER_MODULE
	)

	r1, _, _ := syscall.SyscallN(
		procGetExtendedUdpTable.Addr(),
		uintptr(unsafe.Pointer(unsafe.SliceData(table))),
		uintptr(unsafe.Pointer(size)),
		intbool(order),
		uintptr(syscall.AF_INET),
		UDP_TABLE_OWNER_PID,
		0,
	)
	if r1 == windows.NO_ERROR {
		return nil
	}
	return errors.WithStack(syscall.Errno(r1))
}

const SIO_RCVALL = windows.IOC_IN | windows.IOC_VENDOR | 1

type recvall = uint32

const (
	RCVALL_OFF             recvall = 0
	RCVALL_ON              recvall = 1
	RCVALL_SOCKETLEVELONLY recvall = 2
	RCVALL_IPLEVEL         recvall = 3
)

// https://learn.microsoft.com/en-us/windows/win32/winsock/sio-rcvall
type Recvall struct {
	fd windows.Handle
}

func NewRecvall(laddr netip.Addr, proto int, recvall recvall) (*Recvall, error) {
	switch proto {
	case windows.IPPROTO_IP, windows.IPPROTO_IPV6:
	default:
		return nil, errors.Errorf("invalid ip protocol %d", proto)
	}
	switch recvall {
	case RCVALL_OFF, RCVALL_ON, RCVALL_SOCKETLEVELONLY, RCVALL_IPLEVEL:
	default:
		return nil, errors.Errorf("invalid recvall value %d", recvall)
	}

	fd, err := windows.Socket(windows.AF_INET, windows.SOCK_RAW, int(proto))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if laddr.Is4() && !laddr.IsUnspecified() {
		err = errors.WithStack(
			windows.Bind(fd, &windows.SockaddrInet4{Addr: laddr.As4()}),
		)
	} else if laddr.Is6() && !laddr.IsUnspecified() {
		err = errors.WithStack(
			windows.Bind(fd, &windows.SockaddrInet6{Addr: laddr.As16()}),
		)
	} else {
		err = errors.Errorf("invalid local address %s", laddr.String())
	}
	if err != nil {
		windows.Close(fd)
		return nil, err
	}

	// set SIO_RCVALL io control, on windows, net.IPConn only can read packet than not match
	// any socket.
	if err := windows.WSAIoctl(
		fd,                                // descriptor identifying a socket
		SIO_RCVALL,                        // dwIoControlCode
		(*byte)(unsafe.Pointer(&recvall)), // lpvInBuffer
		uint32(unsafe.Sizeof(recvall)),    // cbInBuffer
		nil,                               // lpvOutBuffer output buffer
		0,                                 // size of output buffer
		new(uint32),                       // number of bytes returned
		nil,                               // OVERLAPPED structure
		0,                                 // completion routine
	); err != nil {
		windows.Close(fd)
		return nil, errors.WithStack(err)
	}

	return &Recvall{fd}, nil
}

func (r *Recvall) Recv(ip []byte) (int, error) {
	var n uint32
	if err := windows.WSARecv(
		r.fd,
		&windows.WSABuf{Len: uint32(len(ip)), Buf: unsafe.SliceData(ip)},
		1,
		&n,
		new(uint32),
		nil,
		nil,
	); err != nil {
		return 0, err // windows.WSAEMSGSIZE
	}
	return int(n), nil
}

func (r *Recvall) Close() error { return errors.WithStack(windows.Close(r.fd)) }

func QueryFullProcessImageNameW(proc windows.Handle, flags uint32, name []uint16, size *uint32) error {
	r1, _, e := syscall.SyscallN(
		procQueryFullProcessImageNameW.Addr(),
		uintptr(proc),
		uintptr(flags),
		uintptr(unsafe.Pointer(unsafe.SliceData(name))),
		uintptr(unsafe.Pointer(size)),
	)
	if r1 == 0 {
		return errors.WithStack(e)
	}
	return nil
}

func GetProcessImageFileNameW(proc windows.Handle, name []uint16) (uint32, error) {
	r1, _, e := syscall.SyscallN(
		procGetProcessImageFileNameW.Addr(),
		uintptr(proc),
		uintptr(unsafe.Pointer(unsafe.SliceData(name))),
		uintptr(len(name)),
	)
	if r1 == 0 {
		return 0, errors.WithStack(e)
	}
	return uint32(r1), nil
}

func QueryDosDeviceW(deviceName string, TargetPath []uint16) error {
	var namePrt uintptr
	if len(deviceName) > 0 {
		n, err := windows.UTF16PtrFromString(deviceName)
		if err != nil {
			return errors.WithStack(err)
		}
		namePrt = uintptr(unsafe.Pointer(n))
	}

	r1, _, e := syscall.SyscallN(
		procQueryDosDeviceW.Addr(),
		namePrt,
		uintptr(unsafe.Pointer(unsafe.SliceData(TargetPath))),
		uintptr(len(TargetPath)),
	)
	if r1 == 0 {
		return errors.WithStack(e)
	}
	return nil
}

const IF_NAMESIZE = 256

// not alias name
func IfIndexToName(ifidx int) (string, error) {
	var b [IF_NAMESIZE]byte
	r1, _, e := syscall.SyscallN(
		procIfIndexToName.Addr(),
		uintptr(ifidx),
		uintptr(unsafe.Pointer(&b[0])),
	)
	if r1 == 0 {
		if e != windows.NOERROR {
			return "", errors.WithStack(e)
		}
		return "", errors.WithStack(windows.ERROR_FILE_NOT_FOUND)
	}

	var n int = IF_NAMESIZE
	for i, e := range b {
		if e == 0 {
			n = i
			break
		}
	}
	var s = make([]byte, n)
	copy(s, b[:])
	return string(s), nil
}

func ConvertInterfaceIndexToLuid(idx uint32) (luid winipcfg.LUID, err error) {
	luid, err = winipcfg.LUIDFromIndex(idx)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return luid, nil
}

func ConvertInterfaceLuidToAlias(luid uint64) (string, error) {
	var buff = make([]uint16, 16)
	for {
		r1, _, _ := syscall.SyscallN(
			procConvertInterfaceLuidToAlias.Addr(),
			uintptr(unsafe.Pointer(&luid)),
			uintptr(unsafe.Pointer(&buff[0])),
			uintptr(len(buff)),
		)
		if e := windows.Errno(r1); e != windows.NO_ERROR {
			if e == windows.ERROR_NOT_ENOUGH_MEMORY {
				buff = append(buff, 0)
				buff = buff[:cap(buff)]
				continue
			}
			return "", errors.WithStack(e)
		}

		return windows.UTF16ToString(buff), nil
	}
}
