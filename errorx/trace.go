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
func Trace(err error) slog.Attr {
	pc := ConcatTraceAndCallers(err, 1)
	return slog.Attr{Key: "trace", Value: slog.AnyValue(pc)}
}

type Frames []uintptr

func (t Frames) LogValue() slog.Value {
	var attrs []slog.Attr
	fs := runtime.CallersFrames(t)
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

func ConcatTraceAndCallers(err error, callSkip int) []uintptr {
	p := getPc()
	defer putPc(p)
	pc := unsafe.Slice((*uintptr)(unsafe.Pointer(p)), size)

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
	//     C:/Go/src/runtime/proc.go:271
	//     C:/Go/src/runtime/asm_amd64.s:1695
	pc = pc[:max(len(pc)-2, 0)]

	errorsPc := pc[:len(pc):cap(pc)]
	callerPc := pc[len(pc):cap(pc)]

	// get call trace stack trace
	callerPc = callerPc[:runtime.Callers(2+callSkip, callerPc)]
	callerPc = callerPc[:max(len(callerPc)-2, 0)]

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
