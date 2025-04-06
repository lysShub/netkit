package errorx

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

type CloseErr struct {
	err atomic.Pointer[error]
}

func (c *CloseErr) Close(fn func() (errs []error)) error {
	if c.err.CompareAndSwap(nil, &_emptyErr) {
		var err error
		if fn != nil {
			for _, e := range fn() {
				if e != nil && err == nil {
					err = e
					break
				}
			}
		}

		if err != nil {
			c.err.Store(&err)
			return err
		} else {
			err = errors.WithStack(net.ErrClosed)
			c.err.Store(&err)
			return nil // not err
		}
	} else {
		return c.Error()
	}
}

// Error 返回Close时的err, 没有Close时返回nil
func (c *CloseErr) Error() (err error) {
start:
	if e := c.err.Load(); e == &_emptyErr {
		time.Sleep(time.Millisecond)
		goto start
	} else if e != nil {
		err = *e
	}
	return err
}
func (c *CloseErr) Closed() bool { return c.err.Load() != nil }
