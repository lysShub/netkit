package errorx_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/lysShub/netkit/errorx"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func Test_CloseErr(t *testing.T) {
	{
		var closeErr errorx.CloseErr

		err := closeErr.Close(nil)
		require.NoError(t, err)

		err = closeErr.Close(nil)
		require.True(t, errors.Is(err, net.ErrClosed))
	}

	{
		var closeErr errorx.CloseErr

		err := closeErr.Close(func() (errs []error) {
			errs = append(errs, errors.New("1234"))
			return
		})
		require.Contains(t, err.Error(), "1234")

		err = closeErr.Close(nil)
		require.Contains(t, err.Error(), "1234")
	}

	t.Run("parallel", func(t *testing.T) {
		var eg, _ = errgroup.WithContext(context.Background())

		var closeErr errorx.CloseErr
		for i := 0; i < 0xff; i++ {
			if i%2 == 0 {
				eg.Go(func() error {
					err := closeErr.Close(func() (errs []error) {
						errs = append(errs, errors.New("1234"))
						return
					})
					require.Equal(t, err.Error(), "1234")
					return nil
				})
			} else {
				eg.Go(func() error {
					err := closeErr.Close(nil)
					require.Equal(t, err.Error(), "1234")
					return nil
				})
			}
		}
		eg.Wait()
	})
}
