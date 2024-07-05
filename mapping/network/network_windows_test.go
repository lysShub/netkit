package network

import (
	"fmt"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Network(t *testing.T) {
	m := New()

	require.NoError(t, m.Upgrade(0))

	m.Query(syscall.IPPROTO_TCP, func(e Elem) (stop bool) {
		fmt.Println(e.Laddr.String(), filepath.Base(e.Path))
		return false
	})

}
