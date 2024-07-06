package pcap

import (
	"os"
	"testing"

	"github.com/google/gopacket/pcapgo"
	"github.com/stretchr/testify/require"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

func Test_Pcap(t *testing.T) {
	var file = "test.pcap"
	os.Remove(file)

	var eth = header.Ethernet{
		0x72, 0x99, 0x96, 0x10, 0x34, 0x2a, 0x80, 0x64, 0x64, 0x18, 0x77, 0x6f, 0x08, 0x00, 0x45, 0x00,
		0x00, 0x4a, 0xeb, 0x85, 0x40, 0x00, 0x80, 0x06, 0x3e, 0x78, 0xc0, 0xa8, 0x2b, 0x23, 0x72, 0x72,
		0x72, 0x72, 0xfb, 0x03, 0x00, 0x35, 0x25, 0x38, 0x0f, 0x7f, 0xbc, 0x07, 0x04, 0x85, 0x50, 0x18,
		0xfa, 0xf0, 0x02, 0xe7, 0x00, 0x00, 0xfd, 0x87, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x03, 0x61, 0x70, 0x69, 0x08, 0x62, 0x69, 0x6c, 0x69, 0x62, 0x69, 0x6c, 0x69, 0x03,
		0x63, 0x6f, 0x6d, 0x00, 0x00, 0x41, 0x00, 0x01,
	}
	func() {
		p, err := File(file)
		require.NoError(t, err)
		defer p.Close()

		err = p.Write(eth)
		require.NoError(t, err)
	}()

	func() {
		p, err := File(file)
		require.NoError(t, err)
		defer p.Close()

		err = p.WriteIP(eth[header.EthernetMinimumSize:])
		require.NoError(t, err)
	}()

	fh, err := os.Open(file)
	require.NoError(t, err)
	defer fh.Close()
	r, err := pcapgo.NewReader(fh)
	require.NoError(t, err)

	{
		data, _, err := r.ReadPacketData()
		require.NoError(t, err)
		require.Equal(t, []byte(eth), data)
	}
	{
		data, _, err := r.ReadPacketData()
		require.NoError(t, err)
		require.Equal(t, eth.Type(), header.Ethernet(data).Type())
		require.Equal(t, []byte(eth[header.EthernetMinimumSize:]), data[header.EthernetMinimumSize:])
	}
	{
		_, _, err := r.ReadPacketData()
		require.Equal(t, "EOF", err.Error())
	}
}
