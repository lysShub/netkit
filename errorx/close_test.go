package errorx_test

import (
	"context"
	"errors"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/lysShub/netkit/errorx"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func Test_CloseErr(t *testing.T) {
	t.Run("close函数为nil", func(t *testing.T) {
		var closeErr errorx.CloseErr

		err := closeErr.Close(nil)
		require.NoError(t, err)

		err = closeErr.Close(nil)
		require.True(t, errors.Is(err, net.ErrClosed))
	})
	t.Run("close没有发生错误", func(t *testing.T) {
		var closeErr errorx.CloseErr

		err := closeErr.Close(func() (errs []error) { return nil })
		require.NoError(t, err)

		err = closeErr.Close(nil)
		require.True(t, errors.Is(err, net.ErrClosed))
	})

	t.Run("close中发生错误, 只记录第一个错误", func(t *testing.T) {
		var closeErr errorx.CloseErr

		err := closeErr.Close(func() (errs []error) {
			errs = append(errs, errors.New("1234"))
			errs = append(errs, errors.New("5678"))
			return
		})
		require.Contains(t, err.Error(), "1234")
		require.NotContains(t, err.Error(), "5678")

		err = closeErr.Close(nil)
		require.Contains(t, err.Error(), "1234")
		require.NotContains(t, err.Error(), "5678")
	})

	t.Run("closing时调用Error()、Close()", func(t *testing.T) {
		var closeErr errorx.CloseErr

		var wg = &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			closeErr.Close(func() (errs []error) {
				wg.Done()
				time.Sleep(time.Second * 2)
				return nil
			})
		}()

		wg.Wait()
		require.True(t, closeErr.Closed())
		time.Sleep(time.Second)

		start := time.Now()
		require.True(t, errors.Is(closeErr.Error(), net.ErrClosed))
		require.InDelta(t, time.Since(start), time.Second, float64(time.Millisecond*100))
	})

	t.Run("closing时调用Close()", func(t *testing.T) {
		var closeErr errorx.CloseErr

		var wg = &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			closeErr.Close(func() (errs []error) {
				wg.Done()
				time.Sleep(time.Second * 2)
				return nil
			})
		}()

		wg.Wait()

		start := time.Now()
		closeErr.Close(func() (errs []error) {
			errs = append(errs, os.ErrNotExist)
			return errs
		})
		require.InDelta(t, time.Since(start), time.Second*2, float64(time.Millisecond*100))

		require.True(t, errors.Is(closeErr.Error(), net.ErrClosed))
	})

	t.Run("parallel", func(t *testing.T) {
		var eg, _ = errgroup.WithContext(context.Background())

		var closeErr errorx.CloseErr
		for i := 0; i < 0xff; i++ {
			if i%2 == 0 {
				eg.Go(func() error {
					err := closeErr.Close(func() (errs []error) {
						time.Sleep(time.Second)
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
