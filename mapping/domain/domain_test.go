package domain

import (
	"context"
	"fmt"
	"net/netip"
	"os"
	"slices"
	"syscall"
	"testing"

	"github.com/google/gopacket/pcapgo"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

var dnsResps = func(t require.TestingT) (resps [][]byte) {
	fh, err := os.Open("test.pcap")
	require.NoError(t, err)
	defer fh.Close()

	r, err := pcapgo.NewReader(fh)
	require.NoError(t, err)

	for {
		data, _, err := r.ReadPacketData()
		if err != nil && err.Error() == "EOF" {
			break
		}
		require.NoError(t, err)

		ip := header.IPv4(data[14:])
		if ip.Protocol() == syscall.IPPROTO_UDP {
			udp := header.UDP(ip.Payload())
			resps = append(resps, slices.Clone(udp.Payload()))
		}
	}
	return resps
}

func Test_Cache(t *testing.T) {
	c, err := NewCache()
	require.NoError(t, err)
	defer c.Close()

	for _, e := range dnsResps(t) {
		c.putRaw(e)
	}

	t.Run("cname", func(t *testing.T) {
		names := c.RDNS(netip.MustParseAddr("52.109.52.131"))
		require.Equal(t, []string{
			"jpe-azsc-config.officeapps.live.com",
			"asia.configsvc1.live.com.akadns.net",
			"prod.configsvc1.live.com.akadns.net",
			"config.officeapps.live.com",
			"officeclient.microsoft.com",
		}, names)
	})

	t.Run("normal", func(t *testing.T) {
		var addrs = make([]netip.Addr, 0)
		{
			var m dns.Msg
			m.SetQuestion(dns.Fqdn("live.bilibili.com"), dns.TypeA)
			r, err := dns.ExchangeContext(context.Background(), &m, "114.114.114.114:53")
			require.NoError(t, err)
			for _, e := range r.Answer {
				if e, ok := (e).(*dns.A); ok {
					addrs = append(addrs, netip.AddrFrom4([4]byte(e.A.To4())))
				}
			}
		}

		for _, e := range addrs {
			str := e.String()
			fmt.Println(str)
			names := c.RDNS(e)
			require.Contains(t, names, "live.bilibili.com")
		}
	})
}
