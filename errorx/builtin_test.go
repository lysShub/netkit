package errorx

import (
	"net"
	"net/http"
	"testing"
	"time"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func Test_Temporary(t *testing.T) {
	var e1 = errors.WithMessage(&net.DNSError{IsTemporary: true}, "temp")
	require.True(t, Temporary(e1))

	var e2 = errors.WithStack(errors.New("error"))
	require.False(t, Temporary(e2))

	var e3 error = nil
	require.False(t, Temporary(e3))
}

func Test_Timeout(t *testing.T) {
	var e1 = errors.WithMessage(&net.DNSError{IsTimeout: true}, "temp")
	require.True(t, Timeout(e1))

	var e2 = errors.WithStack(errors.New("error"))
	require.False(t, Timeout(e2))

	var e3 error = nil
	require.False(t, Timeout(e3))
}

func Test_ConnectRefused(t *testing.T) {
	_, err := (&http.Client{Timeout: time.Second}).Get(`http://localhost:12345`)
	require.Error(t, err)
	require.True(t, ConnectRefused(err))
}

func Test_Builtin(t *testing.T) {
	require.True(t, ConnectRefused(ErrConnectRefused))
	require.True(t, NetworkUnreach(ErrNetworkUnreach))
	require.True(t, BuffTooSmall(ErrBuffTooSmall))
}

func Test_EmptyErr(t *testing.T) {
	require.NotEqual(t, uintptr(unsafe.Pointer(&ErrEmpty)), uintptr(unsafe.Pointer(&_emptyErr)))
	require.False(t, _emptyErr == nil)
	require.False(t, ErrEmpty == nil)
}
