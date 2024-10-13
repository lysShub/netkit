package errorx

import (
	"net"
	"testing"

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
