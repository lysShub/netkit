package errorx

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func Test_Timeout(t *testing.T) {
	var e1 = errors.WithMessage(context.DeadlineExceeded, "temp")
	require.True(t, Timeout(e1))

	var e2 = errors.WithStack(errors.New("error"))
	require.False(t, Timeout(e2))

	var e3 error = nil
	require.False(t, Timeout(e3))
}
