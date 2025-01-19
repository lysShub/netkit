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
		for {
			if c.err.Load() == &_emptyErr {
				runtime.Gosched()
				continue
			}
			return (*c.err.Load())
		}
	}
}

func (c *CloseErr) Closed() bool { return c.err.Load() != nil }
