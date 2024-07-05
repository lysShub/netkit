package domain

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXxxx(t *testing.T) {

	// , &net.IPAddr{IP: net.IP{114, 114, 114, 114}}

	conn, err := net.ListenIP("ip4:udp", &net.IPAddr{IP: net.IP{192, 168, 43, 35}})
	require.NoError(t, err)
	defer conn.Close()

	var b = make([]byte, 1536)
	for {
		n, raddr, err := conn.ReadFrom(b)
		require.NoError(t, err)

		fmt.Println(n, raddr.String())
	}

}
