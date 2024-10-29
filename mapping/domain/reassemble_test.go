package domain

import (
	"os"
	"syscall"
	"testing"

	"github.com/google/gopacket/pcapgo"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

func Test_TcpAssembler(t *testing.T) {
	a := NewTcpAssembler()

	fh, err := os.Open(`./test.pcap`)
	require.NoError(t, err)
	defer fh.Close()

	r, err := pcapgo.NewReader(fh)
	require.NoError(t, err)

	var count int
	for {
		data, info, err := r.ReadPacketData()
		if err != nil {
			if err.Error() == "EOF" {
				break
			} else {
				require.NoError(t, err)
			}
		}

		hdr := header.IPv4(data[14:])
		if hdr.Protocol() == syscall.IPPROTO_TCP {
			data, err = a.Put(hdr, info.Timestamp)
			require.NoError(t, err)
			if len(data) > 0 {
				msgs := RawDnsOverTcp(data).Msgs()
				for _, msg := range msgs {
					require.NoError(t, (&dns.Msg{}).Unpack(msg))
					count++
				}
			}
		}
	}
	require.Greater(t, count, 0)
	require.Zero(t, len(a.flows))
}
