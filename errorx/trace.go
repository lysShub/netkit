package errorx

import (
	"fmt"
	"log/slog"
	"runtime"
	"slices"
	"strconv"

	"github.com/pkg/errors"
)

// Trace trans github.com/pkg/errors stack trace as slog.Attr
//
// Example:
//
//	slog.Error(err.Error(), errorx.Trace(err))
func Trace(err error) slog.Attr {
	var pcs []uintptr

	// get err with stack trace, only hit innermost
	type trace interface{ StackTrace() errors.StackTrace }
	var t trace
	for e := err; e != nil; {
		if e1, ok := e.(trace); ok {
			t = e1
		}
		e = errors.Unwrap(e)
	}
	if t != nil {
		for _, e := range t.StackTrace() {
			pcs = append(pcs, uintptr(e))
		}
	}

	// get call Trace stack trace
	var pc2 = make([]uintptr, 16)
	pc2 = pc2[:runtime.Callers(2, pc2)]
	if len(pcs) > 1 {
		for j, pc := range pc2 {
			i := slices.Index(pcs, pc)
			if i >= 0 {
				pcs = append(pcs[:i], pc2[0])
				pcs = append(pcs, pc2[j:]...)
				break
			}
		}
	} else {
		pcs = append(pcs, pc2...)
	}

	var attrs []slog.Attr
	fs := runtime.CallersFrames(pcs[:max(len(pcs)-2, 0)])
	for {
		f, more := fs.Next()

		attrs = append(attrs, slog.Attr{
			Key:   strconv.Itoa(len(attrs)),
			Value: slog.StringValue(fmt.Sprintf("%s:%d", f.File, f.Line)),
		})
		if !more {
			break
		}
	}

	return slog.Attr{Key: "trace", Value: slog.GroupValue(attrs...)}
}
