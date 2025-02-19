package grpcerr

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_FromError(t *testing.T) {
	msg := "token已过期"
	var err = status.Error(codes.InvalidArgument, msg)
	err = errors.WithStack(err)

	s, ok := FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, s.Code())
	require.Equal(t, msg, s.Message())
}
