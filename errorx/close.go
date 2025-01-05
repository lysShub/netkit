package errorx

import (
	"net"
	"runtime"
	"sync/atomic"

	"github.com/pkg/errors"
)

type CloseErr struct {
	err atomic.Pointer[error]
}

func (c *CloseErr) Close(fn func() (errs []error)) (err error) {
	if c.err.CompareAndSwap(nil, &_emptyErr) {
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
		} else {
			e := errors.WithStack(net.ErrClosed)
			c.err.Store(&e)
		}
		return err
	} else {
		for {
			p := c.err.Load()
			if p == &_emptyErr {
				runtime.Gosched()
				continue
			}
			return (*p)
		}
	}
}

func (c *CloseErr) Closed() bool { return c.err.Load() != nil }
