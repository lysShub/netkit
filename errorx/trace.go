package errorx

import (
	"log/slog"
	"runtime"
	"slices"
	"strconv"
	"sync"
	"unsafe"
	_ "unsafe"

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

		var b = make([]byte, 0, len(f.File)+7)
		b = append(b, f.File...)
		b = append(b, ':')
		b = strconv.AppendInt(b, int64(f.Line), 10)

		attrs = append(attrs, slog.Attr{
			Key:   strconv.Itoa(len(attrs)),
			Value: slog.StringValue(unsafe.String(unsafe.SliceData(b), len(b))),
		})
		if !more {
			break
		}
	}
	return slog.GroupValue(attrs...)
}

//go:linkname appendEscapedJSONString log/slog.appendEscapedJSONString
func appendEscapedJSONString(buf []byte, s string) []byte

func (f Frames) AppendJSON(b []byte) ([]byte, error) {
	if len(f) == 0 {
		b = append(b, '{', '}')
		return b, nil
	}
	fs := runtime.CallersFrames(f)

	b = append(b, '{')
	for i := 0; ; i++ {
		f, more := fs.Next()

		b = append(b, '"')
		b = strconv.AppendInt(b, int64(i), 10)
		b = append(b, '"', ':', '"')
		b = appendEscapedJSONString(b, f.File)
		b = append(b, ':')
		b = strconv.AppendInt(b, int64(f.Line), 10)
		b = append(b, '"', ',')

		if !more {
			break
		}
	}
	b[len(b)-1] = '}'

	return b, nil
}

func (f Frames) MarshalJSON() ([]byte, error) {
	return f.AppendJSON(make([]byte, 0, 256))
}

type trace interface{ StackTrace() errors.StackTrace }

func WithTrace(err error) bool {
	for e := err; e != nil; {
		if _, ok := e.(trace); ok {
			return true
		}
		e = errors.Unwrap(e)
	}
	return false
}

//go:noinline
func ConcatTraceAndCallers(err error, callSkip int) (fs Frames) {
	pcPtr := pcPool.Get().(*[]uintptr)
	pc := *pcPtr
	defer func() {
		*pcPtr = pc[:0]
		pcPool.Put(pcPtr)
	}()

	// get err with stack trace, only hit innermost
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
	pc = (pc)[:max(len(pc)-1, 0)]

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

var pcPool = sync.Pool{
	New: func() any {
		var pc = make([]uintptr, 0, 64)
		return &pc
	},
}
