//go:build windows
// +build windows

package domain

import (
	"fmt"
	"net/netip"
	"os"
	"slices"
	"syscall"
	"testing"

	"github.com/google/gopacket/pcapgo"
	"github.com/lysShub/divert-go"
	"github.com/lysShub/netkit/errorx"
	"github.com/stretchr/testify/require"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

var dnsMsgs = func(t require.TestingT) (msgs [][]byte) {
	fh, err := os.Open("test.pcap") // todo: 这个数据集不完美、没有一次tcp查询多个记录的
	require.NoError(t, err)
	defer fh.Close()

	r, err := pcapgo.NewReader(fh)
	require.NoError(t, err)

	a := NewTcpAssembler()
	for {
		data, info, err := r.ReadPacketData()
		if err != nil && err.Error() == "EOF" {
			break
		}
		require.NoError(t, err)

		ip := header.IPv4(data[14:])
		switch ip.Protocol() {
		case syscall.IPPROTO_UDP:
			udp := header.UDP(ip.Payload())
			msgs = append(msgs, slices.Clone(udp.Payload()))
		case syscall.IPPROTO_TCP:
			data, err := a.Put(ip, info.Timestamp)
			require.NoError(t, err)

			if len(data) > 0 {
				for _, msg := range RawDnsOverTcp(data).Msgs() {
					msgs = append(msgs, slices.Clone(msg))
				}
			}
		case 1:
		default:
			t.Errorf("invalid protocol %d", ip.Protocol())
			t.FailNow()
		}
	}
	return msgs
}

func Test_Cache(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	defer c.Close()
	for _, e := range dnsMsgs(t) {
		require.NoError(t, c.put(e))
	}

	t.Run("udp-cname", func(t *testing.T) {
		for _, addr := range []netip.Addr{
			netip.MustParseAddr("40.126.35.19"),
			netip.MustParseAddr("40.126.35.145"),
			netip.MustParseAddr("20.190.163.19"),
		} {
			names := c.RDNS(addr)
			require.Equal(t, []string{
				"www.tm.v4.a.prd.aadg.trafficmanager.net",
				"prdv4a.aadg.msidentity.com",
				"www.tm.lg.prod.aadmsa.trafficmanager.net",
				"login.msa.msidentity.com",
				"login.live.com",
			}, names)
		}
	})

	t.Run("tcp-cname", func(t *testing.T) {
		for _, addr := range []netip.Addr{
			netip.MustParseAddr("118.180.40.35"),
			netip.MustParseAddr("125.74.1.35"),
		} {
			names := c.RDNS(addr)
			require.Equal(t, []string{
				"opencdnbilibiliv6.jomodns.com",
				"i0.hdslb.com.a.bdydns.com",
				"i0.hdslb.com",
			}, names)
		}
	})

}

func TestXxxx(t *testing.T) {
	var b = make([]byte, 1536)

	divert.MustLoad(divert.DLL)
	d, err := divert.Open("outbound and !loopback and ip and tcp", divert.Network, 0, divert.ReadOnly|divert.Sniff)
	require.NoError(t, err)

	var dup = map[netip.Addr]bool{}
	for {
		n, err := d.Recv(b[:cap(b)], nil)
		require.NoError(t, err)

		dst := netip.AddrFrom4(header.IPv4(b[:n]).DestinationAddress().As4())
		var has, ok bool
		if ok, has = dup[dst]; ok {
			continue
		}

		names, err := RDNS(dst)
		require.True(t, err == nil || errorx.Temporary(err), err)

		dup[dst] = len(names) > 0
		if !has {
			fmt.Println(dst.String(), names)
		}
	}
}
