package errorx_test

import (
	"errors"
	"io"
	"testing"

	"github.com/lysShub/netkit/errorx"
	"github.com/stretchr/testify/require"
)

func Test_ShortBuff(t *testing.T) {
	err := errorx.ShortBuff(1, 2)

	require.True(t, errors.Is(err, io.ErrShortBuffer))
	require.True(t, errorx.Temporary(err))
}
