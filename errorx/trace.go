package errorx

import (
	"fmt"
	"log/slog"
	"runtime"
	"slices"
	"strconv"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
)

// Trace trans github.com/pkg/errors stack trace as slog.Attr
//
// Example:
//
//	slog.Error(err.Error(), errorx.Trace(err))
//
//go:noinline
func Trace(err error) slog.Attr {
	pc := ConcatTraceAndCallers(err, 1)
	return slog.Attr{Key: "trace", Value: slog.AnyValue(pc)}
}

type Frames []uintptr

func (f Frames) LogValue() slog.Value {
	if len(f) == 0 {
		return slog.GroupValue()
	}

	var attrs []slog.Attr
	fs := runtime.CallersFrames(f)
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
	return slog.GroupValue(attrs...)
}

func (f Frames) JsonLiteral(dst []byte) ([]byte, error) {
	if len(f) == 0 {
		return dst, nil
	}

	dst = append(dst, '{')

	fs := runtime.CallersFrames(f)
	for i := 0; ; i++ {
		f, more := fs.Next()

		dst = append(dst, '"')
		dst = strconv.AppendInt(dst, int64(i), 10)
		dst = append(dst, '"', ':', '"')
		dst = append(dst, f.File...)
		dst = append(dst, ':')
		dst = strconv.AppendInt(dst, int64(f.Line), 10)
		dst = append(dst, '"', ',')

		if !more {
			break
		}
	}

	dst[len(dst)-1] = '}'

	return dst, nil
}

//go:noinline
func ConcatTraceAndCallers(err error, callSkip int) Frames {
	p := getPc()
	defer putPc(p)
	pc := unsafe.Slice((*uintptr)(unsafe.Pointer(p)), size)[:0]

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
			pc = append(pc, uintptr(e))
		}
	}
	// remove runtime position, like:
	//     C:/Go/src/runtime/asm_amd64.s:1695
	pc = pc[:max(len(pc)-1, 0)]

	errorsPc := pc[:len(pc):cap(pc)]
	callerPc := pc[len(pc):cap(pc)]

	// get call trace stack trace
	callerPc = callerPc[:runtime.Callers(2+callSkip, callerPc)]
	callerPc = callerPc[:max(len(callerPc)-1, 0)]

	var i int
	for j, p := range callerPc {
		i = slices.Index(errorsPc, p)
		if i >= 0 {
			// add self caller positon
			errorsPc = append(errorsPc[:i], callerPc[0])

			// append caller trace
			errorsPc = append(errorsPc, callerPc[j:]...)
			break
		}
	}
	if i < 0 {
		return Frames(slices.Clone(pc[:len(errorsPc)+len(callerPc)]))
	} else {
		return Frames(slices.Clone(errorsPc))
	}
}

const size = 64

var pcPool = sync.Pool{
	New: func() any {
		return &([size]uintptr{})
	},
}

func getPc() *[size]uintptr {
	return pcPool.Get().(*[size]uintptr)
}
func putPc(p *[size]uintptr) {
	if p != nil {
		pcPool.Put(p)
	}
}
