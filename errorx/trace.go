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
	var pc = make([]uintptr, 36)
	pc = ConcatTraceAndCallers(err, 1, pc)

	var attrs []slog.Attr
	fs := runtime.CallersFrames(pc)
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

func ConcatTraceAndCallers(err error, callSkip int, pc []uintptr) []uintptr {
	pc = pc[:0]

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
	epc := pc[:len(pc):cap(pc)] // err trace stack
	cpc := pc[len(pc):cap(pc)]  // caller stack

	// get call trace stack trace
	cpc = cpc[:runtime.Callers(2+callSkip, cpc)]

	// remove runtime position, like:
	//     C:/Go/src/runtime/proc.go:271
	//     C:/Go/src/runtime/asm_amd64.s:1695
	cpc = cpc[:max(len(cpc)-2, 0)]

	var i int
	for j, p := range cpc {
		i = slices.Index(epc, p)
		if i >= 0 {
			// add self caller positon
			epc = append(epc[:i], cpc[0])

			// append caller trace
			epc = append(epc, cpc[j:]...)
			break
		}
	}
	if i < 0 {
		return pc[:len(epc)+len(cpc)]
	} else {
		return epc
	}
}

/*
func position(p uintptr) string {
	pc := uintptr(p) - 1
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "not func"
	}
	file, line := fn.FileLine(pc)
	// file = relpath(file)
	strn := strconv.Itoa(line)

	b := strings.Builder{}
	b.Grow(len(file) + 1 + len(strn))
	b.WriteString(file)
	b.WriteRune(':')
	b.WriteString(strn)
	return b.String()
}
*/
